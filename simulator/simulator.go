package main

import (
	"fmt"
	"io"
	"bufio"
	"log"
	"os"
	"dubcc/datatypes"
)

func instructionFromWord(
	s *datatypes.Sim,
	word datatypes.MachineWord,
) (datatypes.Instruction, bool) {
	word &= ^uint16(datatypes.InstImmediateFlag)
	word &= ^uint16(datatypes.InstIndirectAFlag)
	word &= ^uint16(datatypes.InstIndirectBFlag)
	inst, found := s.Isa.MOT[word]
	return inst, found
}

func main() {
	memCap := datatypes.MachineAddress(1<<5)
	sim := datatypes.MakeSim(memCap) // 32b for now
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
	var instWord datatypes.MachineWord
	var inst datatypes.Instruction
	args := make([]datatypes.MachineWord, 2)
	buf := make([]byte, 2) // read one words worth at a time
	var v datatypes.MachineWord
	outer_loop: for {
		if (state != Exec) {
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
			v = datatypes.MachineWord(buf[0] << 8 + buf[1])
			fmt.Fprintf(os.Stderr, "got word %x (%d) out of %v\n", v, v, buf)
		}

		switch state {
			case GetOp:
				instWord = v
				inst_, ifound := instructionFromWord(&sim, instWord)
				if !ifound {
					fmt.Fprintf(os.Stderr,
						"invalid instruction %x (%d)\n", instWord, instWord,
					)
				}
				inst = inst_
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
				inst, hfound := sim.Handlers[inst.Repr]
				if !hfound {
					fmt.Fprintf(os.Stderr, "couldn't handle instruction %v\n", inst)
				}
				state = GetOp
		}
	}
	log.Printf("Simulation state: %#v", sim)
}
