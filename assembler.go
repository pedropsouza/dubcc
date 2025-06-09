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
)

type (
	MachineAddress = uint64
	MachineWord = uint16
)
type DirectiveHandler struct {
	f func(info *Info, line InLine)
	numArgs int
}

type Info struct {
	isa ISA
	directives map[string]DirectiveHandler
	symbols map[string]MachineAddress
	undefSyms UndefSymChain
	output []MachineWord
	line_counter MachineAddress
}

type ISA struct {
	instructions map[string]Instruction
}

type Instruction struct {
	numArgs int
	repr MachineWord
}

type UndefSymChainLink struct {
	addr MachineAddress // address for the link data in the binary
	prev MachineAddress // != 0 if this link is not the last for this symbol
	from MachineAddress // address of the unresolved code pos
	sign byte   // FIXME: iunno what this one does
	name string
}

type UndefSymChain struct {
	links []UndefSymChainLink
	top MachineAddress
	base MachineAddress
}

func (usymchain *UndefSymChain) ChainSym(
	from MachineAddress,
	name string,
) *UndefSymChainLink {
	var prevlink *UndefSymChainLink = nil
	for _, link := range slices.Backward(usymchain.links) {
		if link.name == name {
			prevlink = &link
			break;
		}
	}
	
	prev := MachineAddress(0)
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
	out MachineWord
}

func (info *Info) firstPass(line InLine) (reprs []Repr, err error) {
	isa := info.isa
	idata, ifound := isa.instructions[line.op]
	if !ifound {
		// try the directives
		directive, dfound := info.directives[line.op]
		if !dfound {
			log.Fatal("invalid operation: %v", line.op)
		}
		directive.f(info, line)
		return nil, nil
	}
	r := make([]Repr, 1+idata.numArgs)
	if int(idata.numArgs) != len(line.args) {
		return nil, errors.New("number of arguments doesn't match")
	}
	r[0] = Repr {
		tag: ReprComplete,
		input: line.op,
		out: idata.repr,
	}
	
	for index, arg := range line.args {
		index += 1
		// try constant interpretation
		r[index].input = arg
		num, err := parseNum(arg)
		if err == nil {
			r[index].tag = ReprComplete
			r[index].symbol = arg
			r[index].out = MachineWord(num) // will overflow, panic maybe?
			continue
		}
		// aight it ain't a number
		// check symbol table
		lookup, found := info.symbols[arg]
		if found {
			r[index].tag = ReprComplete
			r[index].symbol = arg
			r[index].out = MachineWord(lookup)
		} else {
			// new link should be added
			from := MachineAddress(len(info.output) + 1 + index)
			newLink := info.undefSyms.ChainSym(from, arg)
			r[index].tag = ReprPartial
			r[index].symbol = arg
			r[index].out = MachineWord(newLink.addr)
		}
	}

	for _, repr := range r {
		info.output = append(info.output, repr.out)
	}

	info.line_counter = MachineAddress(len(info.output))
	
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

func (info *Info) registerConst(name string, val MachineWord) {
	if name != "" {
		info.symbols[name] = info.line_counter
	}
	info.output = append(info.output, val)
	info.line_counter += 1
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	info := Info {
		isa: ISA {
			map[string]Instruction {
				"add":   Instruction { numArgs: 1, repr: 2 },
				"br":    Instruction { numArgs: 1, repr: 0 },
				"brneg": Instruction { numArgs: 1, repr: 5 },
				"brpos": Instruction { numArgs: 1, repr: 1 },
				"brzero":Instruction { numArgs: 1, repr: 4 },
				"copy":  Instruction { numArgs: 2, repr: 13 },
				"divide":Instruction { numArgs: 1, repr: 10 },
				"load":  Instruction { numArgs: 1, repr: 3 },
				"mult":  Instruction { numArgs: 1, repr: 14 },
				"read":  Instruction { numArgs: 1, repr: 12 },
				"stop":  Instruction { numArgs: 0, repr: 11 },
				"store": Instruction { numArgs: 1, repr: 7 },
				"sub":   Instruction { numArgs: 1, repr: 6 },
				"write": Instruction { numArgs: 1, repr: 8 },
			},
		},
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
					info.registerConst(line.label, MachineWord(num))
				},
				numArgs: 1,
			},
		},
		symbols: make(map[string]MachineAddress),
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
			info.output[link.from] = MachineWord(sym)
		}
	}
	fmt.Fprintf(os.Stderr, "%#v\n", info)
	fmt.Print(info.output)
}
