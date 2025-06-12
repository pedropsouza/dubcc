package main

import (
	"io"
	"bufio"
	"log"
	"os"
	"dubcc/datatypes"
)

type InstHandler func (*Sim, []datatypes.MachineWord)

type Sim struct {
	mem SimMem
	handlers map[datatypes.MachineWord]InstHandler
	isa datatypes.ISA
}

type SimMem struct {
	work []datatypes.MachineWord
}

func isImmediate(op datatypes.MachineWord) bool {
	return (op & datatypes.InstImmediateFlag) != 0
}

func isIndirectA(op datatypes.MachineWord) bool {
	return (op & datatypes.InstIndirectAFlag) != 0
}

func isIndirectB(op datatypes.MachineWord) bool {
	return (op & datatypes.InstIndirectBFlag) != 0
}

func mapRegisterUnary(
	reg *datatypes.Register,
	mapfunc func (datatypes.MachineWord, datatypes.MachineWord) datatypes.MachineWord,
  value datatypes.MachineWord) {
	reg.Content = mapfunc(reg.Content, value)
}

func resolveAddressMode(s *Sim, opword datatypes.MachineWord, args []datatypes.MachineWord) (a datatypes.MachineWord, b datatypes.MachineWord) {
	a = args[0]
	hasb := len(args) > 1
	if hasb {
		b = args[1]
	}
	if isImmediate(opword) {
		return a, 0 // no immediate binary instructions i believe
	} else {
		if isIndirectA(opword) {
			a = s.mem.work[s.mem.work[a]]
		}
		if hasb && isIndirectB(opword) {
			b = s.mem.work[s.mem.work[b]]
		} else {
			a = s.mem.work[a]
			if hasb {
				b = s.mem.work[b]
			}
		}
		return a, b
	}
}

func makeSim(memSize datatypes.MachineAddress) Sim {
	instHandlers := map[string]InstHandler {
		"add": func (s *Sim, args []datatypes.MachineWord) {
			opword := args[0]
			value, _ := resolveAddressMode(s, opword, args[1:])
			mapRegisterUnary(s.isa.Registers["ACC"],
				func (acc, val datatypes.MachineWord) datatypes.MachineWord {
					return acc + val
				},
				value,
			)
		},
		"sub": func (s *Sim, args []datatypes.MachineWord) {
			opword := args[0]
			value, _ := resolveAddressMode(s, opword, args[1:])
			mapRegisterUnary(s.isa.Registers["ACC"],
				func (acc, val datatypes.MachineWord) datatypes.MachineWord {
					return acc - val
				},
				value,
			)
		},
	}

	isa := datatypes.GetDefaultISA()
	mopHandlers := make(map[datatypes.MachineWord]InstHandler)
	for name, handler := range instHandlers {
		mopHandlers[isa.Instructions[name].Repr] = handler
	}

	return Sim {
		mem: SimMem {
			work: make([]datatypes.MachineWord, memSize),
		},
		isa: isa,
		handlers: mopHandlers,
	}
}

func main() {
	memCap := datatypes.MachineAddress(1<<5)
	sim := makeSim(memCap) // 32b for now
	log.Printf("loaded %d", memCap)
	reader := bufio.NewReader(os.Stdin)

	const (
		Exec = iota
		GetTwoArgs1
		GetTwoArgs2
		GetSingleArg
		GetOp
	)

	state := GetOp
	var inst datatypes.Instruction
	args := make([]datatypes.MachineWord, 2)
	buf := make([]byte, 2) // read one words worth at a time
	var v datatypes.MachineWord
	outer_loop: for {
		for idx, _ := range buf {
			readb, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					break outer_loop
				}
				log.Fatal("error reading stdin: %v", err)
			}
			buf[idx] = readb
		}

		v = datatypes.MachineWord(buf[1] << 8 + buf[0])
		switch state {
			case GetOp:
				inst = sim.isa.MOT[v]
				state = inst.NumArgs
			case GetTwoArgs1:
				state -= 1
				args[0] = v
			case GetTwoArgs2:
				state -= 1
				args[1] = v
			case GetSingleArg:
				state = Exec
				args[0] = v
			case Exec:
				log.Printf("Executing %s with %v", inst.Name, args)
				sim.handlers[inst.Repr](&sim, args)
		}
	}
	log.Printf("Simulation state: %#v", sim)
}
