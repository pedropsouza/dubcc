package assembler

import (
	"dubcc"
	"errors"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"fmt"
	"github.com/k0kubun/pp/v3"
)

var (
	globalSymbols = make(map[string]bool)
	externSymbols = make(map[string]bool)
	maxStackSize  *dubcc.MachineAddress
	moduleEnded   bool
)

type Info struct {
	isa              dubcc.ISA
	directives       map[string]DirectiveHandler
	symbols          map[string]dubcc.MachineAddress
	symbolOccurances map[string][]dubcc.MachineAddress
	undefSyms        UndefSymChain
	macros           map[string]Macros
	macroLevel       int
	macroStack       []MacroFrame
	output           []dubcc.MachineWord
	lineCounter     dubcc.MachineAddress
	StartAddress     dubcc.MachineAddress
	stackSize		     dubcc.MachineAddress
	moduleEnded      bool
}

func (info *Info) GetOutput() []dubcc.MachineWord {
	return info.output
}

type DirectiveHandler struct {
	f       func(info *Info, line dubcc.InLine)
	numArgs int
}

type UndefSymChainLink struct {
	addr dubcc.MachineAddress // address for the link data in the binary
	prev dubcc.MachineAddress // != 0 if this link is not the last for this symbol
	from dubcc.MachineAddress // address of the unresolved code pos
	sign byte                 // FIXME: iunno what this one does
	name string
}

type UndefSymChain struct {
	links []UndefSymChainLink
	top   dubcc.MachineAddress
	base  dubcc.MachineAddress
}

type Macros struct {
	args      []string
	body      []string
	definedAt dubcc.MachineAddress //Totalmente opcional, uso futuro para mensagens de erro
}

type MacroFrame struct {
	name string
	args []string
	body []string
}

func BoolToInt(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

// Lida com os símbolos indefinidos
func (usymchain *UndefSymChain) ChainSym(
	from dubcc.MachineAddress,
	name string,
) *UndefSymChainLink {
	var prevlink *UndefSymChainLink = nil
	for _, link := range slices.Backward(usymchain.links) {
		if link.name == name {
			prevlink = &link
			break
		}
	}

	prev := dubcc.MachineAddress(0)
	if prevlink != nil {
		prev = prevlink.addr
	}

	newLink := UndefSymChainLink{
		addr: usymchain.top,
		prev: prev,
		from: from,
		sign: byte('+'),
		name: name,
	}
	usymchain.links = append(usymchain.links, newLink)
	usymchain.top += 8 + 8 + 8 + 1 // u64 + u64 + u64 + byte
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

type ReprKind uint8

const (
	ReprRaw ReprKind = iota
	ReprPartial
	ReprComplete
)

type Repr struct {
	tag    ReprKind          //Estado da representação
	input  string            //Texto de entrada
	symbol string            //Nome do símbolo
	out    dubcc.MachineWord //Representação binária
}

func (info *Info) FirstPassString(rawLine string) (reprs []Repr, err error) {
	line := strings.TrimSpace(rawLine)
	parsedLine, err := dubcc.ParseAsmLine(line)
	if err != nil {
		return nil, err
	}
	return info.FirstPass(parsedLine)
}

func (info *Info) FirstPass(line dubcc.InLine) ([]Repr, error) {
	idata, ifound := info.isa.Instructions[line.Op]
	if ifound { //Try instruction
		return info.handleInstruction(line, idata)
	}

	directive, dfound := info.directives[line.Op]
	if dfound { //Try the directive
		directive.f(info, line)
		return nil, nil
	}

	log.Printf("Warning: Invalid operation: %v", line.Op)
	return nil, nil
}

func (info *Info) handleInstruction(line dubcc.InLine, idata dubcc.Instruction) ([]Repr, error) {
	r := make([]Repr, 1+idata.NumArgs)
	if int(idata.NumArgs) != len(line.Args) {
		return nil, errors.New("number of arguments doesn't match")
	}
	r[0] = Repr{
		tag:   ReprComplete,
		input: line.Op,
		out:   idata.Repr,
	}

	for index, arg := range line.Args {
		index += 1
		repr := &r[index]

		// try constant interpretation
		repr.input = arg
		num, err := ParseNum(arg)
		if err == nil {
			repr.tag = ReprComplete
			repr.symbol = arg
			// can overflow, panic maybe?
			// + immediate flag
			repr.out = dubcc.MachineWord(num)
			r[0].out |= dubcc.OpImmediateFlag
			continue
		}
		// aight it ain't a number
		// check if it's a register
		{
			regflag := dubcc.MachineWord(dubcc.OpRegAFlag * BoolToInt(index == 1))
			regflag |= dubcc.MachineWord(dubcc.OpRegBFlag * BoolToInt(index == 2))
			reg, found := info.isa.Registers[arg]
			if found {
				repr.tag = ReprComplete
				repr.symbol = arg
				repr.out = dubcc.MachineWord(reg.Address)
				r[0].out |= regflag
				continue
			}
		}
		{ // check symbol table
			lookup, found := info.symbols[arg]
			if found {
				repr.tag = ReprComplete
				repr.symbol = arg
				repr.out = dubcc.MachineWord(lookup)
			} else {
				// new link should be added
				from := dubcc.MachineAddress(len(info.output) + index)
				newLink := info.undefSyms.ChainSym(from, arg)
				repr.tag = ReprPartial
				repr.symbol = arg
				repr.out = dubcc.MachineWord(newLink.addr)
			}
			info.symbolOccurances[arg] = append(
				info.symbolOccurances[arg],
				dubcc.MachineAddress(len(info.output) + index),
			)
		}
		// TODO/FIXME: decide which
		// syntax we should use to
		// signify indirect mode and implement it
	}

	if line.Label != "" {
		info.registerLabel(line.Label)
	}

	for _, repr := range r {
		pp.Fprintf(os.Stderr, "adding %v @ %v\n", repr.out, len(info.output))
		info.output = append(info.output, repr.out)
	}
	
	info.lineCounter = dubcc.MachineAddress(len(info.output))

	return r, nil
}

func (info *Info) SecondPass() map[string]dubcc.MachineAddress {
	for _, link := range info.undefSyms.links {
		sym, found := info.symbols[link.name]
		if !found {
			if !IsExternalSymbol(link.name) {
				log.Printf("%s", fmt.Sprint(globalSymbols))
				log.Fatalf("undefined symbol: %v (%v)", link.name, link)
			}
		}
		info.output[link.from] = dubcc.MachineWord(sym)
	}
	return info.symbols
}

func ParseNum(in string) (num uint64, err error) {
	b2 := regexp.MustCompile("^0b([0-1]+)$")
	b8 := regexp.MustCompile("^0o([0-7]+)$")
	b10 := regexp.MustCompile("^([0-9]+)$")
	b16 := regexp.MustCompile("^0x([0-9abcdefABCDEF]+)$")

	recognizerBaseMap := map[*regexp.Regexp]int{
		b2:  2,
		b8:  8,
		b10: 10,
		b16: 16,
	}
	for recognizer, base := range recognizerBaseMap {
		matches := recognizer.FindStringSubmatch(in)
		if len(matches) > 1 {
			match := matches[1]
			num, err := strconv.ParseInt(match, base, 64)
			if err != nil {
				return 0, err
			}
			return uint64(num), nil
		}
	}
	return 0, errors.New("invalid number")
}

func (info *Info) registerLabelAt(name string, where dubcc.MachineAddress) {
	info.symbols[name] = where
}

func (info *Info) registerLabel(name string) {
	info.registerLabelAt(name, info.lineCounter)
}

func (info *Info) GetSymbols() []string {
	var syms []string
	for sym := range info.symbols {
		syms = append(syms, sym)
	}
	return syms
}

func (info *Info) SymbolOccurances() map[string][]dubcc.MachineAddress {
	return info.symbolOccurances
}

func (info *Info) registerConst(name string, val dubcc.MachineWord) {
	if name != "" {
		info.symbols[name] = info.lineCounter
	}
	info.output = append(info.output, val)
	info.lineCounter += 1
}

type ObjectInfo struct {
	globalSymbols  map[string]bool
	sections       map[string]*Section
	currentSection *Section
}

var globalObjectInfoMap map[*Info]*ObjectInfo

func MakeAssembler() Info {
  // when assembling more than 1 file
	if globalObjectInfoMap == nil {
		globalObjectInfoMap = make(map[*Info]*ObjectInfo)
	}
	
	info := Info{
		isa:        dubcc.GetDefaultISA(),
		directives: Directives(),
		symbols:    make(map[string]dubcc.MachineAddress),
		symbolOccurances: make(map[string][]dubcc.MachineAddress),
		macros:     make(map[string]Macros),
	}
	
	return info
}

func Directives() map[string]DirectiveHandler {
	return map[string]DirectiveHandler{
		"space": {
			f: func(info *Info, line dubcc.InLine) {
				info.registerConst(line.Label, 0)
			},
			numArgs: 0,
		},
		"const": {
			f: func(info *Info, line dubcc.InLine) {
				num, err := ParseNum(line.Args[0])
				if err != nil {
					log.Fatalf("can't decide value for const %v: %v", line.Label, err)
				}
				info.registerConst(line.Label, dubcc.MachineWord(num))
			},
			numArgs: 1,
		},
		"end": {
			f: func(info *Info, line dubcc.InLine) {
				info.moduleEnded = true
				moduleEnded = true
				log.Printf("module ended at address 0x%x", info.lineCounter)
			},
			numArgs: 0,
		},
		"extr": {
			f: func(info *Info, line dubcc.InLine) {
				symbolName := line.Args[0]
				globalSymbols[symbolName] = true
				log.Printf("declared global symbol: %s", symbolName)
				if addr, exists := info.symbols[symbolName]; exists {
					log.Printf("symbol %s already defined at 0x%x, marking as global", symbolName, addr)
				}
			},
			numArgs: 1,
		},
		"extdef": {
			f: func(info *Info, line dubcc.InLine) {
				if line.Label == "" {
					log.Fatalf("extdef requires a label, got %s", line.Label)
				}
				externSymbols[line.Label] = true
				log.Printf("declared external symbol: %s", line.Label)
			},
			numArgs: 0,
		},
		"stack": {
			f: func(info *Info, line dubcc.InLine) {
				num, err := ParseNum(line.Args[0])
				if err != nil {
					log.Fatalf("can't parse stack size %v: %v", line.Args[0], err)
				}
				stackSize := dubcc.MachineAddress(num)
				maxStackSize = &stackSize
				log.Printf("set maximum stack size to %d words", stackSize)
			},
			numArgs: 1,
		},
		"start": {
			f: func(info *Info, line dubcc.InLine) {
				if len(line.Args) != 1 {
					log.Fatalf("start directive requires one argument (address), got %d", len(line.Args))
				}
				addrStr := line.Args[0]
				addr, err := ParseNum(addrStr)
				if err != nil {
					log.Fatalf("invalid start address: %v", err)
				}
				info.StartAddress = dubcc.MachineAddress(addr)
			},
			numArgs: 1,
		},
		"MACRO": {
			f:       func(info *Info, line dubcc.InLine) {},
			numArgs: 0,
		},
		"MEND": {
			f:       func(info *Info, line dubcc.InLine) {},
			numArgs: 0,
		},
	}
}

func IsGlobalSymbol(name string) bool {
	return globalSymbols[name]
}

func IsExternalSymbol(name string) bool {
	return externSymbols[name]
}
