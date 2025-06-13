package datatypes

type (
	MachineAddress = uint64
	MachineWord = uint16
)

type ISA struct {
	Instructions map[string]Instruction
	Registers map[string]*Register
	MOT map[MachineWord]Instruction
}

type Instruction struct {
	Name string
	NumArgs int
	Repr MachineWord
}

const (
	InstIndirectAFlag = (1 << 5) << iota
	InstIndirectBFlag
	InstImmediateFlag
)

type RegisterTag byte
const (
	RegisterTagGeneralPurpose = 1 << iota
	RegisterTagSpecial
	RegisterTagInternal
)

type Register struct {
	Name string
	Desc string
	Size uint
	Longdesc string
	Tags RegisterTag
	Content MachineWord
}

func GetDefaultISA () ISA {
	insts := map[string]Instruction {
		"add":   Instruction { Name: "add", NumArgs: 1, Repr: 2 },
		"br":    Instruction { Name: "br", NumArgs: 1, Repr: 0 },
		"brneg": Instruction { Name: "brneg", NumArgs: 1, Repr: 5 },
		"brpos": Instruction { Name: "brpos", NumArgs: 1, Repr: 1 },
		"brzero":Instruction { Name: "brzero", NumArgs: 1, Repr: 4 },
		"copy":  Instruction { Name: "copy", NumArgs: 2, Repr: 13 },
		"divide":Instruction { Name: "divide", NumArgs: 1, Repr: 10 },
		"load":  Instruction { Name: "load", NumArgs: 1, Repr: 3 },
		"mult":  Instruction { Name: "mult", NumArgs: 1, Repr: 14 },
		"read":  Instruction { Name: "read", NumArgs: 1, Repr: 12 },
		"stop":  Instruction { Name: "stop", NumArgs: 0, Repr: 11 },
		"store": Instruction { Name: "store", NumArgs: 1, Repr: 7 },
		"sub":   Instruction { Name: "sub", NumArgs: 1, Repr: 6 },
		"write": Instruction { Name: "write", NumArgs: 1, Repr: 8 },
	}

	mot := make(map[MachineWord]Instruction)
	for _, inst := range insts {
		mot[inst.Repr] = inst
	}

	return ISA {
		Instructions: insts,
		MOT: mot,
		Registers: map[string]*Register {
			"PC": &Register {
				Name: "PC", Desc: "Contador de Instruções (Program Counter)", Size: 16,
				Longdesc: "Mantém o endereço da próxima instrução a ser executada",
				Tags: RegisterTagSpecial,
			},
			"SP": &Register {
				Name: "SP", Desc: "Ponteiro de pilha (Stack Pointer)", Size: 16,
				Longdesc: "Aponta para o topo da pilha do sistema; tem incremento/decremento automático (push/pop)",
				Tags: RegisterTagSpecial,
			},
			"ACC": &Register {
				Name: "ACC", Desc: "Acumulador", Size: 16,
				Longdesc: "Armazena os dados (carregados e resultantes) das operações da Unid. de Lógica e Aritmética",
				Tags: RegisterTagGeneralPurpose | RegisterTagSpecial,
			},
			"MOP": &Register {
				Name: "MOP", Desc: "Modo de Operação", Size: 8,
				Longdesc: "Armazena o indicador do modo de operação, que é alterado apenas por painel de operação (via console de operação - interface visual)",
				Tags: RegisterTagSpecial,
			},
			"RI": &Register {
				Name: "RI", Desc: "Registrador de Instrução", Size: 16,
				Longdesc: "Mantém o opcode da instrução em execução (registrador interno)",
				Tags: RegisterTagSpecial | RegisterTagInternal,
			},
			"RE": &Register {
				Name: "RE", Desc: "Registrador de Endereço de Memória", Size: 16,
				Longdesc: "Mantém o endereço de acesso à memória de dados (registrador interno)",
				Tags: RegisterTagSpecial | RegisterTagInternal,
			},
			"R0": &Register {
				Name: "R0", Desc: "Registrador de Propósito Geral", Size: 16,
				Longdesc: "Registrador de Propósito Geral",
				Tags: RegisterTagGeneralPurpose,
			},
			"R1": &Register {
				Name: "R1", Desc: "Registrador de Propósito Geral", Size: 16,
				Longdesc: "Registrador de Propósito Geral",
				Tags: RegisterTagGeneralPurpose,
			},
		},
	}
}

type InstHandler func (*Sim, []MachineWord)

type Sim struct {
	Mem SimMem
	Handlers map[MachineWord]InstHandler
	Isa ISA
}

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
	mapfunc func (MachineWord, MachineWord) MachineWord,
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
	instHandlers := map[string]InstHandler {
		"add": func (s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func (acc, val MachineWord) MachineWord {
					return acc + val
				},
				value,
			)
		},
		"sub": func (s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Isa.Registers["ACC"].MapUnary(
				func (acc, val MachineWord) MachineWord {
					return acc - val
				},
				value,
			)
		},
	}

	isa := GetDefaultISA()
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
