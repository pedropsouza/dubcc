package dubcc

type Instruction struct {
	Name string
	NumArgs int
	Repr MachineWord
	Flags InstructionFlag
}

// static flags time
type InstructionFlag byte
const (
	InstFlagImmediate = 1 << iota // Accepts Immediate values
	InstFlagImmediateB
	InstFlagStack
)

// runtime flags
const (
	InstIndirectAFlag = (1 << 5) << iota
	InstIndirectBFlag
	InstRegAFlag
	InstRegBFlag
	InstImmediateFlag
)

func (sim *Sim) InstructionFromWord(
	word MachineWord,
) (Instruction, bool) {
	word &= 0x001f; // get base repr w/o flags
	inst, found := sim.MOT[word]
	return inst, found
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

type inst = Instruction // shorthand for these defs
func InstMap() map[string]Instruction {
	return map[string]Instruction {
		"add":   inst { Name: "add",   NumArgs: 1, Repr: 2, Flags: InstFlagImmediate },
		"br":    inst { Name: "br",    NumArgs: 1, Repr: 0, Flags: 0 },
		"brneg": inst { Name: "brneg", NumArgs: 1, Repr: 5, Flags: 0 },
		"brpos": inst { Name: "brpos", NumArgs: 1, Repr: 1, Flags: 0 },
		"brzero":inst { Name: "brzero",NumArgs: 1, Repr: 4, Flags: 0 },
		"copy":  inst { Name: "copy",  NumArgs: 2, Repr: 13,Flags: InstFlagImmediateB },
		"divide":inst { Name: "divide",NumArgs: 1, Repr: 10,Flags: InstFlagImmediate },
		"load":  inst { Name: "load",  NumArgs: 1, Repr: 3, Flags: InstFlagImmediate },
		"mult":  inst { Name: "mult",  NumArgs: 1, Repr: 14,Flags: InstFlagImmediate },
		"read":  inst { Name: "read",  NumArgs: 1, Repr: 12,Flags: 0 },
		"ret":   inst { Name: "ret",   NumArgs: 0, Repr: 16,Flags: InstFlagStack },
		"stop":  inst { Name: "stop",  NumArgs: 0, Repr: 11,Flags: 0 },
		"store": inst { Name: "store", NumArgs: 1, Repr: 7, Flags: 0 },
		"sub":   inst { Name: "sub",   NumArgs: 1, Repr: 6, Flags: InstFlagImmediate },
		"write": inst { Name: "write", NumArgs: 1, Repr: 8, Flags: InstFlagImmediate },
	}
}

func registerMap1Handler(
	regAddress MachineAddress,
	mapf func (*Sim, MachineWord, MachineWord) MachineWord,
) InstHandler {
	return func (s *Sim, args []MachineWord) {
		opword := args[0]
		vals := s.ResolveAddressMode(opword, args[1:])
		reg := &s.Registers[regAddress]
		*reg = mapf(s, *reg, *vals[0])
	}
}

func registerMap2Handler(
	regAddress MachineAddress,
	mapf func (*Sim, MachineWord, MachineWord, MachineWord) MachineWord,
) InstHandler {
	return func (s *Sim, args []MachineWord) {
		opword := args[0]
		vals := s.ResolveAddressMode(opword, args[1:])
		reg := &s.Registers[regAddress]
		*reg = mapf(s, *reg, *vals[0], *vals[1])
	}
}

func mutateState1Handler(callback func (*Sim, *MachineWord)) InstHandler {
	return func (s *Sim, args[]MachineWord) {
		opword := args[0]
		vals := s.ResolveAddressMode(opword, args[1:])
		callback(s, vals[0])
	}
}

func mutateState2Handler(callback func (*Sim, *MachineWord, *MachineWord)) InstHandler {
	return func (s *Sim, args[]MachineWord) {
		opword := args[0]
		vals := s.ResolveAddressMode(opword, args[1:])
		callback(s, vals[0], vals[1])
	}
}

func InstHandlers() map[string]InstHandler {
	return map[string]InstHandler {
		"add": registerMap1Handler(
			RegACC,
			func (s *Sim, acc MachineWord, value MachineWord) MachineWord {
				return acc + value
			},
		),
		"sub": registerMap1Handler(
			RegACC,
			func (s *Sim, acc MachineWord, value MachineWord) MachineWord {
				return acc - value
			},
		),
		"divide": registerMap1Handler(
			RegACC,
			func(s *Sim, acc MachineWord, value MachineWord) MachineWord {
				return acc / value
			},
		),
		"mult": registerMap1Handler(
			RegACC,
			func(s *Sim, acc MachineWord, value MachineWord) MachineWord {
				return acc * value
			},
		),
		"br": registerMap1Handler(
			RegPC,
			func (s *Sim, pc MachineWord, value MachineWord) MachineWord {
				return value
			},
		),
		"brpos": registerMap1Handler(
			RegPC,
			func (s *Sim, pc MachineWord, value MachineWord) MachineWord {
				acc_v := s.GetRegister(RegACC)
				if acc_v != 0 && (acc_v & 0x8000) == 0 {
					return value
				} else {
					return pc
				}
			},
		),
		"brneg": registerMap1Handler(
			RegPC,
			func (s *Sim, pc MachineWord, value MachineWord) MachineWord {
				if (s.GetRegister(RegACC) & 0x8000) > 0 {
					return value
				} else {
					return pc
				}
			},
		),
		"brzero": registerMap1Handler(
			RegPC,
			func (s *Sim, pc MachineWord, value MachineWord) MachineWord {
				if s.GetRegister(RegACC) == 0 {
					return value
				} else {
					return pc
				}
			},
		),
		"load": registerMap1Handler(
			RegACC,
			func(s *Sim, acc MachineWord, value MachineWord) MachineWord {
				return value
			},
		),
		"store": mutateState1Handler(func (s *Sim, value *MachineWord) {
				*value = s.GetRegister(RegACC)
			},
		),
		"stop": mutateState1Handler(func (s *Sim, value *MachineWord) {
				s.State = SimStateHalt
			},
		),
		"copy": mutateState2Handler(func (s *Sim, l, r *MachineWord) {
				*l = *r
			},
		),
		"push": mutateState1Handler(func(s *Sim, value *MachineWord) {
				// FIXME: unimplemented
			},
		),
	}
}
