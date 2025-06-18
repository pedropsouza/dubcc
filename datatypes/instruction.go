package datatypes

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
