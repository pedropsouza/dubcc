package datatypes

type (
	MachineAddress = uint64
	MachineWord = uint16
)

type ISA struct {
	Instructions map[string]Instruction
	Registers map[string]*Register
}

type Instruction struct {
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
	return ISA {
		Instructions: map[string]Instruction {
			"add":   Instruction { NumArgs: 1, Repr: 2 },
			"br":    Instruction { NumArgs: 1, Repr: 0 },
			"brneg": Instruction { NumArgs: 1, Repr: 5 },
			"brpos": Instruction { NumArgs: 1, Repr: 1 },
			"brzero":Instruction { NumArgs: 1, Repr: 4 },
			"copy":  Instruction { NumArgs: 2, Repr: 13 },
			"divide":Instruction { NumArgs: 1, Repr: 10 },
			"load":  Instruction { NumArgs: 1, Repr: 3 },
			"mult":  Instruction { NumArgs: 1, Repr: 14 },
			"read":  Instruction { NumArgs: 1, Repr: 12 },
			"stop":  Instruction { NumArgs: 0, Repr: 11 },
			"store": Instruction { NumArgs: 1, Repr: 7 },
			"sub":   Instruction { NumArgs: 1, Repr: 6 },
			"write": Instruction { NumArgs: 1, Repr: 8 },
		},
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
