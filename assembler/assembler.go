package main

import (
	"fmt"
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

func parseAsmLine(rawLine string) (line InLine, err error) {
	label, code, labeled := strings.Cut(rawLine, ":")
	if !labeled {
		code = label
	}
	fields := strings.Fields(code)
	return InLine {
		raw: rawLine,
		label: label,
		op: fields[0],
		args: fields[1:],
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
		// try constant interpretation
		r[index].input = arg
		num, err := parseNum(arg)
		if err == nil {
			r[index].tag = ReprComplete
			r[index].symbol = arg
			// can overflow, panic maybe?
			// + immediate flag
			r[index].out = datatypes.MachineWord(num)
			r[0].out |= datatypes.InstImmediateFlag
			continue
		}
		// aight it ain't a number
		// check symbol table
		lookup, found := info.symbols[arg]
		if found {
			r[index].tag = ReprComplete
			r[index].symbol = arg
			r[index].out = datatypes.MachineWord(lookup)
		} else {
			// new link should be added
			from := datatypes.MachineAddress(len(info.output) + 1 + index)
			newLink := info.undefSyms.ChainSym(from, arg)
			r[index].tag = ReprPartial
			r[index].symbol = arg
			r[index].out = datatypes.MachineWord(newLink.addr)
		}
		// TODO/FIXME: decide which 
		// syntax we should use to
		// signify indirect mode and implement it
	}

	for _, repr := range r {
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

func (info *Info) registerConst(name string, val datatypes.MachineWord) {
	if name != "" {
		info.symbols[name] = info.line_counter
	}
	info.output = append(info.output, val)
	info.line_counter += 1
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	if len(os.Args) == 2 {
		inputFile, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Println(err)
		}
		r := bytes.NewReader(inputFile)
		reader = bufio.NewReader(r)
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
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		
		if err != nil {
			log.Println(err)
			break
		}

		{
			parsedline, err := parseAsmLine(line)
			fmt.Fprintf(os.Stderr, "processing %#v... ", parsedline)
			if err != nil {
				log.Println(err)
				continue
			}
			outLine, err := info.firstPass(parsedline)
		
			if err != nil {
				log.Println(err)
			}
			fmt.Fprintf(os.Stderr, "%#v\n", outLine)
		}
	}

	{ // Second pass
		for _, link := range info.undefSyms.links {
			sym, found := info.symbols[link.name]
			if !found {
				log.Fatalf("undefined symbol: %v (%#v)", link.name, link)
			}
			info.output[link.from] = datatypes.MachineWord(sym)
		}
	}
	fmt.Fprintf(os.Stderr, "%#v\n", info)
	fmt.Print(info.output)
}
