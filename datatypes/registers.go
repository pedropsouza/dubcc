package datatypes

type Register struct {
	Name string
	Address MachineAddress
	Desc string
	Size uint
	Longdesc string
	Tags RegisterTag
}

type RegisterTag byte
const (
	RegisterTagGeneralPurpose = 1 << iota
	RegisterTagSpecial
	RegisterTagInternal
)

const (
	RegPC = iota
	RegSP
	RegACC
	RegMOP
	RegRI
	RegRE
	RegR0
	RegR1
)

func RegisterInfo() map[string]*Register {
	return map[string]*Register {
		"PC": &Register {
			Name: "PC", Desc: "Contador de Instruções (Program Counter)", Size: 16,
			Address: RegPC,
			Longdesc: "Mantém o endereço da próxima instrução a ser executada",
			Tags: RegisterTagSpecial,
		},
		"SP": &Register {
			Name: "SP", Desc: "Ponteiro de pilha (Stack Pointer)", Size: 16,
			Address: RegSP,
			Longdesc: "Aponta para o topo da pilha do sistema; tem incremento/decremento automático (push/pop)",
			Tags: RegisterTagSpecial,
		},
		"ACC": &Register {
			Name: "ACC", Desc: "Acumulador", Size: 16,
			Address: RegACC,
			Longdesc: "Armazena os dados (carregados e resultantes) das operações da Unid. de Lógica e Aritmética",
			Tags: RegisterTagGeneralPurpose | RegisterTagSpecial,
		},
		"MOP": &Register {
			Name: "MOP", Desc: "Modo de Operação", Size: 8,
			Address: RegMOP,
			Longdesc: "Armazena o indicador do modo de operação, que é alterado apenas por painel de operação (via console de operação - interface visual)",
			Tags: RegisterTagSpecial,
		},
		"RI": &Register {
			Name: "RI", Desc: "Registrador de Instrução", Size: 16,
			Address: RegRI,
			Longdesc: "Mantém o opcode da instrução em execução (registrador interno)",
			Tags: RegisterTagSpecial | RegisterTagInternal,
		},
		"RE": &Register {
			Name: "RE", Desc: "Registrador de Endereço de Memória", Size: 16,
			Address: RegRE,
			Longdesc: "Mantém o endereço de acesso à memória de dados (registrador interno)",
			Tags: RegisterTagSpecial | RegisterTagInternal,
		},
		"R0": &Register {
			Name: "R0", Desc: "Registrador de Propósito Geral", Size: 16,
			Address: RegR0,
			Longdesc: "Registrador de Propósito Geral",
			Tags: RegisterTagGeneralPurpose,
		},
		"R1": &Register {
			Name: "R1", Desc: "Registrador de Propósito Geral", Size: 16,
			Address: RegR1,
			Longdesc: "Registrador de Propósito Geral",
			Tags: RegisterTagGeneralPurpose,
		},
	}
}

func StartupRegisters(isa *ISA, memSize MachineAddress) (out []MachineWord) {
	out = make([]MachineWord, len(isa.Registers))
	out[RegSP] = MachineWord(memSize - 1)
	return out
}

