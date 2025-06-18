package datatypes

type (
	MachineAddress = uint64
	MachineWord    = uint16
)

type ISA struct {
	Instructions map[string]Instruction
	Registers    map[string]*Register
	MOT          map[MachineWord]Instruction
}

type Instruction struct {
	Name    string
	NumArgs int
	Repr    MachineWord
}

const {
	InstFlagImmediate = 1 << iota // Accepts Immediate values
	InstFlagImmediateB
	InstFlagStack
}

type InstructionFlag byte
const (
	InstIndirectAFlag = (1 << 5) << iota
	InstIndirectBFlag
	InstImmediateFlag
)

type Instruction struct {
	Name string
	NumArgs int
	Repr MachineWord
	Flags InstructionFlag
}

type RegisterTag byte

const (
	RegisterTagGeneralPurpose = 1 << iota
	RegisterTagSpecial
	RegisterTagInternal
)

type Register struct {
	Name     string
	Desc     string
	Size     uint
	Longdesc string
	Tags     RegisterTag
	Content  MachineWord
}

func GetDefaultISA () ISA {
	insts := map[string]Instruction {
		"add":   Instruction { Name: "add", NumArgs: 1, Repr: 2, Flags: InstFlagImmediate },
		"br":    Instruction { Name: "br", NumArgs: 1, Repr: 0, Flags: 0 },
		"brneg": Instruction { Name: "brneg", NumArgs: 1, Repr: 5, Flags: 0 },
		"brpos": Instruction { Name: "brpos", NumArgs: 1, Repr: 1, Flags: 0 },
		"brzero":Instruction { Name: "brzero", NumArgs: 1, Repr: 4, Flags: 0 },
		"copy":  Instruction { Name: "copy", NumArgs: 2, Repr: 13, Flags: InstFlagImmediateB },
		"divide":Instruction { Name: "divide", NumArgs: 1, Repr: 10, Flags: InstFlagImmediate },
		"load":  Instruction { Name: "load", NumArgs: 1, Repr: 3, Flags: InstFlagImmediate },
		"mult":  Instruction { Name: "mult", NumArgs: 1, Repr: 14, Flags: InstFlagImmediate },
		"read":  Instruction { Name: "read", NumArgs: 1, Repr: 12, Flags: 0 },
		"ret":   Instruction { Name: "ret", NumArgs: 0, Repr: 16, Flags: InstFlagStack }
		"stop":  Instruction { Name: "stop", NumArgs: 0, Repr: 11, Flags: 0 },
		"store": Instruction { Name: "store", NumArgs: 1, Repr: 7, Flags: 0 },
		"sub":   Instruction { Name: "sub", NumArgs: 1, Repr: 6, Flags: InstFlagImmediate },
		"write": Instruction { Name: "write", NumArgs: 1, Repr: 8, Flags: InstFlagImmediate },
	}

	mot := make(map[MachineWord]Instruction)
	for _, inst := range insts {
		mot[inst.Repr] = inst
	}

	return ISA{
		Instructions: insts,
		MOT:          mot,
		Registers: map[string]*Register{
			"PC": &Register{
				Name: "PC", Desc: "Contador de Instruções (Program Counter)", Size: 16,
				Longdesc: "Mantém o endereço da próxima instrução a ser executada",
				Tags:     RegisterTagSpecial,
			},
			"SP": &Register{
				Name: "SP", Desc: "Ponteiro de pilha (Stack Pointer)", Size: 16,
				Longdesc: "Aponta para o topo da pilha do sistema; tem incremento/decremento automático (push/pop)",
				Tags:     RegisterTagSpecial,
			},
			"ACC": &Register{
				Name: "ACC", Desc: "Acumulador", Size: 16,
				Longdesc: "Armazena os dados (carregados e resultantes) das operações da Unid. de Lógica e Aritmética",
				Tags:     RegisterTagGeneralPurpose | RegisterTagSpecial,
			},
			"MOP": &Register{
				Name: "MOP", Desc: "Modo de Operação", Size: 8,
				Longdesc: "Armazena o indicador do modo de operação, que é alterado apenas por painel de operação (via console de operação - interface visual)",
				Tags:     RegisterTagSpecial,
			},
			"RI": &Register{
				Name: "RI", Desc: "Registrador de Instrução", Size: 16,
				Longdesc: "Mantém o opcode da instrução em execução (registrador interno)",
				Tags:     RegisterTagSpecial | RegisterTagInternal,
			},
			"RE": &Register{
				Name: "RE", Desc: "Registrador de Endereço de Memória", Size: 16,
				Longdesc: "Mantém o endereço de acesso à memória de dados (registrador interno)",
				Tags:     RegisterTagSpecial | RegisterTagInternal,
			},
			"R0": &Register{
				Name: "R0", Desc: "Registrador de Propósito Geral", Size: 16,
				Longdesc: "Registrador de Propósito Geral",
				Tags:     RegisterTagGeneralPurpose,
			},
			"R1": &Register{
				Name: "R1", Desc: "Registrador de Propósito Geral", Size: 16,
				Longdesc: "Registrador de Propósito Geral",
				Tags:     RegisterTagGeneralPurpose,
			},
		},
	}
}

func (isa *ISA) InstructionFromWord(
	word MachineWord,
) (Instruction, bool) {
	word &= ^uint16(InstImmediateFlag)
	word &= ^uint16(InstIndirectAFlag)
	word &= ^uint16(InstIndirectBFlag)
	inst, found := isa.MOT[word]
	return inst, found
}

type InstHandler func(*Sim, []MachineWord)

type Sim struct {
	Mem      SimMem
	Handlers map[MachineWord]InstHandler
	Isa      ISA
	State    SimState
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

func IsImmediate(op MachineWord) bool {
	return (op & InstImmediateFlag) != 0
}

func IsIndirectA(op MachineWord) bool {
	return (op & InstIndirectAFlag) != 0
}

func IsIndirectB(op MachineWord) bool {
	return (op & InstIndirectBFlag) != 0
}

func (reg *Register) MapUnary(
	mapfunc func(MachineWord, MachineWord) MachineWord,
	value MachineWord) {
	reg.Content = mapfunc(reg.Content, value)
}

func (s *Sim) ResolveAddressMode(opword MachineWord, args []MachineWord) (a MachineWord, b MachineWord) {
	a = args[0]
	hasb := len(args) > 1
	if hasb {
		b = args[1]
	}
	if IsImmediate(opword) {
		return a, 0 // no immediate binary instructions i believe
	} else {
		if IsIndirectA(opword) {
			a = s.Mem.Work[s.Mem.Work[a]]
		}
		if hasb && IsIndirectB(opword) {
			b = s.Mem.Work[s.Mem.Work[b]]
		} else {
			a = s.Mem.Work[a]
			if hasb {
				b = s.Mem.Work[b]
			}
		}
		return a, b
	}
}

func MakeSim(memSize MachineAddress) Sim {
	instHandlers := map[string]InstHandler{
		"add": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord {
					return acc + val
				},
				value,
			)
		},
		"sub": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord {
					return acc - val
				},
				value,
			)
		},
		"divide": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord { return acc / val },
				value,
			)
		},
		"mult": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord { return acc * val },
				value,
			)
		},
		"br": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["PC"].MapUnary(
				func(pc, target MachineWord) MachineWord {
					return target
				},
				value,
			)
		},
		"brpos": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["PC"].MapUnary(
				func(pc, target MachineWord) MachineWord {
					acc_v := s.Isa.Registers["ACC"].Content
					if acc_v != 0 && (acc_v&0x8000) == 0 {
						return target
					} else {
						return pc
					}
				},
				value,
			)
		},
		"brneg": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["PC"].MapUnary(
				func(pc, target MachineWord) MachineWord {
					if (s.Isa.Registers["ACC"].Content & 0x8000) > 0 {
						return target
					} else {
						return pc
					}
				},
				value,
			)
		},
		"brzero": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["PC"].MapUnary(
				func(pc, target MachineWord) MachineWord {
					if s.Isa.Registers["ACC"].Content == 0 {
						return target
					} else {
						return pc
					}
				},
				value,
			)
		},
		"load": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord { return value },
				value,
			)
		},
		"store": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Mem.Work[value] = s.Isa.Registers["ACC"].Content
		},
		"stop": func(s *Sim, args []MachineWord) {
			s.State = SimStateHalt
		},
		"copy": func(s *Sim, args []MachineWord) {
			opword := args[0]
			l, r := s.ResolveAddressMode(opword, args[1:])
			s.Mem.Work[l] = s.Mem.Work[r]
		},
		"push": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func(acc, val MachineWord) MachineWord { return acc / val },
				value,
			)
		},
	}
	isa := GetDefaultISA()
	isa.Registers["SP"].Content = MachineWord(memSize - 1)
	mopHandlers := make(map[MachineWord]InstHandler)
	for name, handler := range instHandlers {
		mopHandlers[isa.Instructions[name].Repr] = handler
	}

	return Sim {
		Mem: SimMem {
			Work: make([]MachineWord, memSize),
		},
		Isa: isa,
		Handlers: mopHandlers,
	}
}
