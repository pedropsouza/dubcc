package datatypes

type (
	MachineAddress = uint64
	MachineWord    = uint16
)

type ISA struct {
	Instructions map[string]Instruction
	Registers map[string]*Register
}

func GetDefaultISA () ISA {
	return ISA {
		Instructions: InstMap(),
		Registers: RegisterInfo(),
	}
}

type InstHandler func (*Sim, []MachineWord)

type Sim struct {
	Mem      SimMem
	Handlers map[MachineWord]InstHandler
	MOT map[MachineWord]Instruction
	Registers []MachineWord
	Isa ISA
	State SimState
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

func (s *Sim) MapRegister(regAddress MachineAddress, mapf func (MachineWord) MachineWord) {
	old := s.Registers[regAddress]
	s.Registers[regAddress] = mapf(old)
}

func MakeSim(memSize MachineAddress) Sim {
	instHandlers := map[string]InstHandler{
		"add": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegACC,
				func (acc MachineWord) MachineWord {
					return acc + value
				},
			)
		},
		"sub": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegACC,
				func (acc MachineWord) MachineWord {
					return acc - value
				},
			)
		},
		"divide": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegACC,
				func(acc MachineWord) MachineWord { return acc / value },
			)
		},
		"mult": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegACC,
				func(acc MachineWord) MachineWord { return acc * value },
			)
		},
		"br": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegPC,
				func (pc MachineWord) MachineWord {
					return value
				},
			)
		},
		"brpos": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegPC,
				func (pc MachineWord) MachineWord {
					acc_v := s.GetRegister(RegACC)
					if acc_v != 0 && (acc_v & 0x8000) == 0 {
						return value
					} else {
						return pc
					}
				},
			)
		},
		"brneg": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegPC,
				func (pc MachineWord) MachineWord {
					if (s.GetRegister(RegACC) & 0x8000) > 0 {
						return value
					} else {
						return pc
					}
				},
			)
		},
		"brzero": func(s *Sim, args []MachineWord) {
			opword := args[0]
			// if not indirect, must be treated as immediate else labels break
			if !(IsIndirectA(opword) || IsIndirectB(opword)) {
				opword |= InstImmediateFlag
			}
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegPC,
				func (pc MachineWord) MachineWord {
					if s.GetRegister(RegACC) == 0 {
						return value
					} else {
						return pc
					}
				},
			)
		},
		"load": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.MapRegister(
				RegACC,
				func(acc MachineWord) MachineWord { return value },
			)
		},
		"store": func(s *Sim, args []MachineWord) {
			opword := args[0]
			value, _ := s.ResolveAddressMode(opword, args[1:])
			s.Mem.Work[value] = s.GetRegister(RegACC)
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
			// FIXME: unimplemented
		},
	}
	isa := GetDefaultISA()
	
	mot := make(map[MachineWord]Instruction)
	for _, inst := range isa.Instructions {
		mot[inst.Repr] = inst
	}

	regs := make([]MachineWord, len(isa.Registers))
	regs[RegSP] = MachineWord(memSize - 1)

	mopHandlers := make(map[MachineWord]InstHandler)
	for name, handler := range instHandlers {
		mopHandlers[isa.Instructions[name].Repr] = handler
	}

	return Sim {
		Mem: SimMem {
			Work: make([]MachineWord, memSize),
		},
		Isa: isa,
		MOT: mot,
		Registers: regs,
		Handlers: mopHandlers,
	}
}
