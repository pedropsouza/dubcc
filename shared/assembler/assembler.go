package assembler

import (
	"dubcc"
	"errors"
	"github.com/k0kubun/pp/v3"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Info struct {
	isa          dubcc.ISA
	directives   map[string]DirectiveHandler
	symbols      map[string]dubcc.MachineAddress
	undefSyms    UndefSymChain
	macros       map[string]Macros
	macroLevel   int
	macroStack   []MacroFrame
	output       []dubcc.MachineWord
	line_counter dubcc.MachineAddress
}

func (info *Info) GetOutput() []dubcc.MachineWord {
	return info.output
}

type DirectiveHandler struct {
	f       func(info *Info, line InLine)
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

type InLine struct {
	raw   string   //Linha original
	label string   //Rótulo
	op    string   //Operação (instrução ou diretiva)
	args  []string //Argumentos
}

var EmptyLineErr = errors.New("empty line")

// Função que recebe a linha em assembly e separa em rótulo, operações/instruções.
func parseAsmLine(rawLine string) (line InLine, err error) {
	label, code, labeled := strings.Cut(rawLine, ":")
	if !labeled {
		code = label
		label = ""
	}
	// ignore comments
	code, _, _ = strings.Cut(code, ";")
	fields := strings.Fields(code)
	if len(fields) < 1 {
		return InLine{}, EmptyLineErr
	}
	return InLine{
		raw:   rawLine,
		label: label,
		op:    fields[0],
		args:  fields[min(len(fields), 1):],
	}, nil
}

type ReprKind uint8

const (
	ReprRaw ReprKind = iota
	ReprPartial
	ReprComplete
)

// Estrutura usada na primeira passagem
type Repr struct {
	tag    ReprKind          //Estado da representação
	input  string            //Texto de entrada
	symbol string            //Nome do símbolo
	out    dubcc.MachineWord //Representação binária
}

func (info *Info) FirstPassString(rawLine string) (reprs []Repr, err error) {
	line := strings.TrimSpace(rawLine)
	parsedLine, err := parseAsmLine(line)
	if err != nil {
		return nil, err
	}
	return info.FirstPass(parsedLine)
}

func (info *Info) FirstPass(line InLine) ([]Repr, error) {
	if line.op == "MACRO" {
		info.macroLevel++
		return nil, nil
	}
	if info.macroLevel > 0 {
		return info.handleMacro(line)
	}
	idata, ifound := info.isa.Instructions[line.op]
	if ifound { //Try instruction
		return info.handleInstruction(line, idata)
	}

	directive, dfound := info.directives[line.op]
	if dfound { //Try the directive
		directive.f(info, line)
		return nil, nil
	}

	macro, mfound := info.macros[line.op]
	if mfound { //This shit has to be a macro, right?
		return info.expandAndRunMacro(macro, line)
	}
	if line.op == "MEND" { //De preferência, deixar como último teste
		return nil, errors.New("End of macro before start.")
	}

	log.Fatal("Invalid operation: %v", line.op)
	return nil, nil

}

func (info *Info) handleInstruction(line InLine, idata dubcc.Instruction) ([]Repr, error) {
	r := make([]Repr, 1+idata.NumArgs)
	if int(idata.NumArgs) != len(line.args) {
		return nil, errors.New("number of arguments doesn't match")
	}
	r[0] = Repr{
		tag:   ReprComplete,
		input: line.op,
		out:   idata.Repr,
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

	info.line_counter = dubcc.MachineAddress(len(info.output))

	return r, nil
}

func (info *Info) handleMacro(line InLine) (reprs []Repr, err error) {
	if len(info.macroStack) < info.macroLevel { //Se for a primeira linha...
		for len(info.macroStack) < info.macroLevel {
			info.macroStack = append(info.macroStack, MacroFrame{})
		}
		info.macroStack[info.macroLevel-1] = MacroFrame{
			name: line.op,
			args: line.args,
			body: []string{},
		}
		return nil, nil
	}

	if line.op == "MEND" { //Aqui, toda macro foi lida e o MEND vai fechar a macro
		info.macroLevel--
		frame := info.macroStack[info.macroLevel]
		macro := Macros{
			args:      frame.args,
			body:      frame.body,
			definedAt: info.line_counter,
		}

		info.macroStack = slices.Delete(info.macroStack, info.macroLevel, info.macroLevel+1)
		info.macros[frame.name] = macro
		return nil, nil
	}
	frame := &info.macroStack[info.macroLevel-1]
	frame.body = append(frame.body, line.raw)
	return nil, nil
}

func (info *Info) SecondPass() map[string]dubcc.MachineAddress {
	for _, link := range info.undefSyms.links {
		sym, found := info.symbols[link.name]
		if !found {
			log.Fatalf("undefined symbol: %v (%v)", link.name, link)
		}
		info.output[link.from] = dubcc.MachineWord(sym)
	}
	return info.symbols
}

func parseNum(in string) (num uint64, err error) {
	b2 := regexp.MustCompile("^0b[0-1]+$")
	b8 := regexp.MustCompile("^0o[0-7]+$")
	b10 := regexp.MustCompile("^[0-9]+$")
	b16 := regexp.MustCompile("^0x[0-9]+$")

	recognizerBaseMap := map[*regexp.Regexp]int{
		b2:  2,
		b8:  8,
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

func (info *Info) registerLabelAt(name string, where dubcc.MachineAddress) {
	info.symbols[name] = where
}

func (info *Info) registerLabel(name string) {
	info.registerLabelAt(name, info.line_counter)
}

func (info *Info) GetSymbols() []string {
	var syms []string
	for sym := range info.symbols {
		syms = append(syms, sym)
	}
	return syms
}

func (info *Info) registerConst(name string, val dubcc.MachineWord) {
	if name != "" {
		info.symbols[name] = info.line_counter
	}
	info.output = append(info.output, val)
	info.line_counter += 1
}

func (info *Info) expandAndRunMacro(macro Macros, line InLine) ([]Repr, error) {
	if len(line.args) != len(macro.args) {
		return nil, errors.New("number of arguments doesn't match")
	}

	substitutions := make(map[string]string)
	for i, formal := range macro.args {
		substitutions[formal] = line.args[i]
	}

	var allReprs []Repr

	for _, raw := range macro.body {

		words := strings.Split(raw, " ")
		for i, word := range words {
			wdata, wfound := substitutions[word]
			if wfound {
				words[i] = wdata
			}
		}
		expanded := strings.Join(words, " ")

		parsedLine, err := parseAsmLine(expanded)
		if err != nil {
			return nil, err
		}
		reprs, err := info.FirstPass(parsedLine)
		if err != nil {
			return nil, err
		}
		allReprs = append(allReprs, reprs...)
	}
	log.Print(allReprs)
	return allReprs, nil
}

func MakeAssembler() Info {
	return Info{
		isa:        dubcc.GetDefaultISA(),
		directives: Directives(),
		symbols:    make(map[string]dubcc.MachineAddress),
		macros:     make(map[string]Macros),
	}
}

func Directives() map[string]DirectiveHandler {
	return map[string]DirectiveHandler{ //TODO adicionar as outras.
		"space": {
			f: func(info *Info, line InLine) {
				info.registerConst(line.label, 0)
			},
			numArgs: 0,
		},
		"const": {
			f: func(info *Info, line InLine) {
				num, err := parseNum(line.args[0])
				if err != nil {
					log.Fatalf("can't decide value for const %v: %v", line.label, err)
				}
				info.registerConst(line.label, dubcc.MachineWord(num))
			},
			numArgs: 1,
		},
	}
}
