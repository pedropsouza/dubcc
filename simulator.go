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

type Sim struct {
	mem SimMem
}

type SimMem struct {
	work []MachineWord
	registers []Register
}

type RegisterTag byte
const (
	RegisterTagGeneralPurpose = 0
	RegisterTagSpecial = 1 << 1
	RegisterTagInternal = 1 << 2
)

type Register struct {
	name string
	desc string
	size uint
	longdesc string
	tags RegisterTag
	content MachineWord
}

func makeSim(memSize MachineAddress) Sim {
	return Sim {
		mem: SimMem {
			work: make([]MachineWord, memSize),
			registers: []Register {
				Register {
					name: "PC", desc: "Contador de Instruções (Program Counter)", size: 16,
					longdesc: "Mantém o endereço da próxima instrução a ser executada",
					tags: RegisterTagSpecial,
				},
				Register: {
					name: "SP", desc: "Ponteiro de pilha (Stack Pointer)", size: 16,
					longdesc: "Aponta para o topo da pilha do sistema; tem incremento/decremento automático (push/pop)",
					tags: RegisterTagSpecial,
				},
				Register {
					name: "ACC", desc: "Acumulador", size: 16,
					longdesc: "Armazena os dados (carregados e resultantes) das operações da Unid. de Lógica e Aritmética",
					tags: RegisterTagGeneralPurpose | RegisterTagSpecial,
				},
				Register {
					name: "MOP", desc: "Modo de Operação", size: 8,
					longdesc: "Armazena o indicador do modo de operação, que é alterado apenas por painel de operação (via console de operação - interface visual)",
					tags: RegisterTagSpecial,
				},
				Register {
					name: "RI", desc: "Registrador de Instrução", size: 16,
					longdesc: "Mantém o opcode da instrução em execução (registrador interno)"
					tags: RegisterTagSpecial | RegisterTagInternal,
				},
				Register {
					name: "RE", desc: "Registrador de Endereço de Memória", size: 16,
					longdesc: "Mantém o endereço de acesso à memória de dados (registrador interno)",
					tags: RegisterTagSpecial | RegisterTagInternal,
				},
				Register {
					name: "R0", desc: "Registrador de Propósito Geral", size: 16,
					longdesc: "Registrador de Propósito Geral",
					tags: RegisterTagGeneralPurpose,
				},
				Register {
					name: "R1", desc: "Registrador de Propósito Geral", size: 16,
					longdesc: "Registrador de Propósito Geral",
					tags: RegisterTagGeneralPurpose,
				},
			},
		},
	}
}
