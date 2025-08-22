package dubcc

import (
	"strings"
)

type (
	MachineAddress = uint64
	MachineWord    = uint16
)

type ISA struct {
	Instructions map[string]Instruction
	Registers    map[string]*Register
}

func GetDefaultISA() ISA {
	return ISA{
		Instructions: InstMap(),
		Registers:    RegisterInfo(),
	}
}

type InstHandler func(*Sim, []MachineWord)

type InLine struct {
	Raw   string   //Linha original
	Label string   //Rótulo
	Op    string   //Operação (instrução ou diretiva)
	Args  []string //Argumentos
}

// Função que recebe a linha em assembly e separa em rótulo, operações/instruções.
func ParseAsmLine(rawLine string) (line InLine, err error) {
	label, code, labeled := strings.Cut(rawLine, ":")
	if !labeled {
		code = label
		label = ""
	}
	// ignore comments
	code, _, _ = strings.Cut(code, ";")
	fields := strings.Fields(code)
	op := ""
	if len(fields) > 0 { op = fields[0] }
	return InLine{
		Raw:   rawLine,
		Label: label,
		Op:    op,
		Args:  fields[min(len(fields), 1):],
	}, nil
}

type Sim struct {
	Mem       SimMem
	Handlers  map[MachineWord]InstHandler
	MOT       map[MachineWord]Instruction
	Registers []MachineWord
	Isa       ISA
	State     SimState
	SaveTemps	bool
	TempDir   string
	inWords   []MachineWord
	outWords  []MachineWord
}

type SimState = byte

const (
	SimStateHalt = iota
	SimStateRun
	SimStatePause
)

type SimMem struct {
	Work []MachineWord
}

func (s *Sim) ResolveAddressMode(opword MachineWord, args []MachineWord) []*MachineWord {
	inst, found := s.InstructionFromWord(opword)
	if !found {
		panic("bad instruction")
	}
	if inst.NumArgs != len(args) {
		panic("argument count mismatch")
	}

	immediateTests := []func() bool{
		func() bool { return (opword&OpImmediateFlag) != 0 && (inst.Flags&InstImmediateA) != 0 },
		func() bool { return (opword&OpImmediateFlag) != 0 && (inst.Flags&InstImmediateB) != 0 },
	}
	indirectTests := []func() bool{
		func() bool { return (opword & OpIndirectAFlag) != 0 },
		func() bool { return (opword & OpIndirectBFlag) != 0 },
	}
	registerTests := []func() bool{
		func() bool { return (opword & OpRegAFlag) != 0 },
		func() bool { return (opword & OpRegBFlag) != 0 },
	}

	out := make([]*MachineWord, 2)
	for idx, arg := range args {
		isIm := immediateTests[idx]()
		isIn := indirectTests[idx]()
		isReg := registerTests[idx]()

		if isReg {
			out[idx] = &s.Registers[arg]
		} else if isIm {
			box := new(MachineWord)
			*box = arg
			out[idx] = box
		} else if isIn {
			out[idx] = &s.Mem.Work[s.Mem.Work[arg]]
		} else { // only direct remaining
			if (inst.Flags & InstDirectIsImmediate) != 0 {
				// botch for the uuuh branch instructions?
				// where Direct is a goddamn alias for Im
				box := new(MachineWord)
				*box = arg
				out[idx] = box
			} else {
				out[idx] = &s.Mem.Work[arg] // direct
			}
		}
	}
	return out
}

func (s *Sim) GetRegister(regAddress MachineAddress) MachineWord {
	return s.Registers[regAddress]
}

func (s *Sim) GetRegisterByName(name string) MachineWord {
	return s.Registers[s.Isa.Registers[name].Address]
}

func (s *Sim) SetRegister(regAddress MachineAddress, value MachineWord) {
	s.Registers[regAddress] = value
}

func (s *Sim) SetRegisterByName(name string, value MachineWord) {
	s.Registers[s.Isa.Registers[name].Address] = value
}

func (s *Sim) MapRegister(regAddress MachineAddress, mapf func(MachineWord) MachineWord) {
	old := s.Registers[regAddress]
	s.Registers[regAddress] = mapf(old)
}

func (sim *Sim) recvWord(w MachineWord) {
	sim.inWords = append(sim.inWords, w)
}

func (sim *Sim) dlvWord() MachineWord {
	sim.outWords = append(sim.outWords, sim.inWords[len(sim.inWords)-1])
	return sim.outWords[len(sim.outWords)-1]
}

func MakeSim(memSize MachineAddress) Sim {
	isa := GetDefaultISA()

	mot := make(map[MachineWord]Instruction)
	for _, inst := range isa.Instructions {
		mot[inst.Repr] = inst
	}

	mopHandlers := make(map[MachineWord]InstHandler)
	for name, handler := range InstHandlers() {
		mopHandlers[isa.Instructions[name].Repr] = handler
	}

	return Sim{
		Mem: SimMem{
			Work: make([]MachineWord, memSize),
		},
		Isa:       isa,
		MOT:       mot,
		Registers: StartupRegisters(&isa, memSize),
		Handlers:  mopHandlers,
	}
}
