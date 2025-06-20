package main

import (
	"bufio"
	"os"
	"log"
	"errors"
	"strconv"
	"strings"
	"regexp"
	"slices"
	"bytes"
	"dubcc/datatypes"
	"github.com/k0kubun/pp/v3"
)

type DirectiveHandler struct {
	f func(info *Info, line InLine)
	numArgs int
}

type Info struct {
	isa datatypes.ISA
	directives map[string]DirectiveHandler
	symbols map[string]datatypes.MachineAddress
	undefSyms UndefSymChain
	output []datatypes.MachineWord
	line_counter datatypes.MachineAddress
}

type UndefSymChainLink struct {
	addr datatypes.MachineAddress // address for the link data in the binary
	prev datatypes.MachineAddress // != 0 if this link is not the last for this symbol
	from datatypes.MachineAddress // address of the unresolved code pos
	sign byte   // FIXME: iunno what this one does
	name string
}

type UndefSymChain struct {
	links []UndefSymChainLink
	top datatypes.MachineAddress
	base datatypes.MachineAddress
}

func BoolToInt(val bool) int { if val { return 1 } else { return 0 } }

func (usymchain *UndefSymChain) ChainSym(
	from datatypes.MachineAddress,
	name string,
) *UndefSymChainLink {
	var prevlink *UndefSymChainLink = nil
	for _, link := range slices.Backward(usymchain.links) {
		if link.name == name {
			prevlink = &link
			break;
		}
	}
	
	prev := datatypes.MachineAddress(0)
	if prevlink != nil {
		prev = prevlink.addr
	}

	newLink := UndefSymChainLink {
		addr: usymchain.top,
		prev: prev,
		from: from,
		sign: byte('+'),
		name: name,
	}
	usymchain.links = append(usymchain.links, newLink)
	usymchain.top += 8+8+8+1 // u64 + u64 + u64 + byte
	return &newLink
}

func (usymchain *UndefSymChain) LookupSym(name string) *UndefSymChainLink {
	idx := slices.IndexFunc(
		usymchain.links,
		func(link UndefSymChainLink) bool {
			return link.name == name
		},
	)

	if idx < 0 {
		return nil
	} else {
		return &usymchain.links[idx]
	}
}

type InLine struct {
	raw string
	label string
	op string
	args []string
}

var emptyLineErr = errors.New("empty line")

func parseAsmLine(rawLine string) (line InLine, err error) {
	label, code, labeled := strings.Cut(rawLine, ":")
	if !labeled {
		code = label
		label = ""
	}
	fields := strings.Fields(code)
	if len(fields) < 1 {
		return InLine {}, emptyLineErr
	}
	return InLine {
		raw: rawLine,
		label: label,
		op: fields[0],
		args: fields[min(len(fields), 1):],
	}, nil
}

type ReprKind uint8
const (
	ReprRaw ReprKind = iota
	ReprPartial
	ReprComplete
)

type Repr struct {
	tag ReprKind
	input string
	symbol string
	out datatypes.MachineWord
}

func (info *Info) firstPass(line InLine) (reprs []Repr, err error) {
	isa := info.isa
	idata, ifound := isa.Instructions[line.op]
	if !ifound {
		// try the directives
		directive, dfound := info.directives[line.op]
		if !dfound {
			log.Fatal("invalid operation: %v", line.op)
		}
		directive.f(info, line)
		return nil, nil
	}
	r := make([]Repr, 1+idata.NumArgs)
	if int(idata.NumArgs) != len(line.args) {
		return nil, errors.New("number of arguments doesn't match")
	}
	r[0] = Repr {
		tag: ReprComplete,
		input: line.op,
		out: idata.Repr,
	}
	
	for index, arg := range line.args {
		index += 1
		repr := &r[index]
		// try constant interpretation
		repr.input = arg
		num, err := parseNum(arg)
		if err == nil {
			repr.tag = ReprComplete
			repr.symbol = arg
			// can overflow, panic maybe?
			// + immediate flag
			repr.out = datatypes.MachineWord(num)
			r[0].out |= datatypes.InstImmediateFlag
			continue
		}
		// aight it ain't a number
		// check if it's a register
		{
			regflag := datatypes.MachineWord(datatypes.InstRegAFlag * BoolToInt(index == 1))
			regflag |= datatypes.MachineWord(datatypes.InstRegBFlag * BoolToInt(index == 2))
			reg, found := info.isa.Registers[arg]
			if found {
				repr.tag = ReprComplete
				repr.symbol = arg
				repr.out = datatypes.MachineWord(reg.Address)
				r[0].out |= regflag
				continue
			}
		}
			{ // check symbol table
			lookup, found := info.symbols[arg]
			if found {
				repr.tag = ReprComplete
				repr.symbol = arg
				repr.out = datatypes.MachineWord(lookup)
			} else {
				// new link should be added
				from := datatypes.MachineAddress(len(info.output) + index)
				newLink := info.undefSyms.ChainSym(from, arg)
				repr.tag = ReprPartial
				repr.symbol = arg
				repr.out = datatypes.MachineWord(newLink.addr)
			}
		}
		// TODO/FIXME: decide which 
		// syntax we should use to
		// signify indirect mode and implement it
	}

	if line.label != "" {
		info.registerLabel(line.label)
	}

	for _, repr := range r {
		pp.Fprintf(os.Stderr, "adding %v @ %v\n", repr.out, len(info.output))
		info.output = append(info.output, repr.out)
	}

	info.line_counter = datatypes.MachineAddress(len(info.output))
	
	return r, nil
}

func parseNum(in string) (num uint64, err error) {
	b2 := regexp.MustCompile("^0b[0-1]+$")
	b8 := regexp.MustCompile("^0o[0-7]+$")
	b10 := regexp.MustCompile("^[0-9]+$")
	b16 := regexp.MustCompile("^0x[0-9]+$")

	recognizerBaseMap := map[*regexp.Regexp]int {
		b2: 2,
		b8: 8,
		b10: 10,
		b16: 16,
	}
	for recognizer, base := range recognizerBaseMap {
		matches := recognizer.Match([]byte(in))
		if matches {
			num, err := strconv.ParseInt(in, base, 64)
			if err != nil {
				return 0, err
			}
			return uint64(num), nil
		}
	}
	return 0, errors.New("invalid number")
}

func (info *Info) registerLabelAt(name string, where datatypes.MachineAddress) {
	info.symbols[name] = where
}

func (info *Info) registerLabel(name string) {
	info.registerLabelAt(name, info.line_counter)
}

func (info *Info) registerConst(name string, val datatypes.MachineWord) {
	if name != "" {
		info.symbols[name] = info.line_counter
	}
	info.output = append(info.output, val)
	info.line_counter += 1
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	if len(os.Args) == 2 {
		inputFile, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Println(err)
		}
		r := bytes.NewReader(inputFile)
		scanner = bufio.NewScanner(r)
	}

	info := Info {
		isa: datatypes.GetDefaultISA(),
		directives: map[string]DirectiveHandler {
			"space": DirectiveHandler {
				f: func (info *Info, line InLine) {
					info.registerConst(line.label, 0)
				},
				numArgs: 0,
			},
			"const": DirectiveHandler {
				f: func (info *Info, line InLine) {
					num, err := parseNum(line.args[0])
					if err != nil {
						log.Fatalf("can't decide value for const %v: %v", line.label, err)
					}
					info.registerConst(line.label, datatypes.MachineWord(num))
				},
				numArgs: 1,
			},
		},
		symbols: make(map[string]datatypes.MachineAddress),
	}
	
	for {
		if !scanner.Scan() { break }
		line := strings.TrimSpace(scanner.Text())
		
		{
			parsedline, err := parseAsmLine(line)
			pp.Fprintf(os.Stderr, "processing %v... ", parsedline)
			if err != nil {
				if err != emptyLineErr {
					log.Println(err)
				}
				continue
			}
			outLine, err := info.firstPass(parsedline)
		
			if err != nil {
				log.Println(err)
			}
			pp.Fprintf(os.Stderr, "%v\n", outLine)
		}
	}

	{ // Second pass
		for _, link := range info.undefSyms.links {
			sym, found := info.symbols[link.name]
			if !found {
				log.Fatalf("undefined symbol: %v (%v)", link.name, link)
			}
			info.output[link.from] = datatypes.MachineWord(sym)
		}
	}
	pp.Fprintf(os.Stderr, "Symbols: %v\n", info.symbols)

	{ // write binary
		writer := bufio.NewWriter(os.Stdout)
		for _, u16 := range info.output {
			high := byte((u16 >> 8) & 0xff)
			low := byte((u16 >> 0) & 0xff)
			writer.WriteByte(high)
			writer.WriteByte(low)
		}
		writer.Flush()
	}
}
