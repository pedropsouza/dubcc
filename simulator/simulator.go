package main

import (
	"fmt"
	"io"
	"bufio"
	"log"
	"os"
	"os/signal"
	"time"
	"dubcc/datatypes"
	"github.com/k0kubun/pp/v3"
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

	{ // install interrupt handler
		c := make (chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for _ = range c {
				sim.State = datatypes.SimStateHalt
			}
		}()
	}

	{ // read bin to memory
		buf := make([]byte, 2) // read one words worth at a time
		read_file: for mempos := range sim.Mem.Work {
			for idx := range buf {
				readb, err := reader.ReadByte()
				if err != nil {
					if err == io.EOF {
						break read_file
					}
					log.Fatal("error reading stdin: %v", err)
				}
				buf[idx] = readb
			}
			v := datatypes.MachineWord(buf[0] << 8 + buf[1])
			fmt.Fprintf(os.Stderr, "got word %x (%d) out of %v\n", v, v, buf)
			sim.Mem.Work[mempos] = v
		}
	}

	sim.State = datatypes.SimStateRun
	for {
		if sim.State != datatypes.SimStateRun {
			break
		}
		pc := sim.Isa.Registers["PC"]
		ri := sim.Isa.Registers["RI"]
		//re := sim.Isa.Registers["RE"]
		ri.Content = sim.Mem.Work[pc.Content]
		inst, ifound := instructionFromWord(&sim, ri.Content)
		if !ifound {
			fmt.Fprintf(os.Stderr,
				"invalid instruction %x (%d)\n", ri.Content, ri.Content,
			)
		}
		handler, hfound := sim.Handlers[inst.Repr]
		if !hfound {
			fmt.Fprintf(os.Stderr, "couldn't handle instruction %v\n", inst)
		}
		instPos := datatypes.MachineAddress(pc.Content)
		argsTerm := instPos + 1 + datatypes.MachineAddress(inst.NumArgs)
		// set pc before calling the handler
		// that way branching works
		pc.Content += datatypes.MachineWord(1 + inst.NumArgs)
		args := sim.Mem.Work[instPos:argsTerm]
		log.Printf("Executing %s with %v", inst.Name, args)
		handler(&sim, args)
		time.Sleep(100 * time.Millisecond)
	}
	pp.Printf("Simulation state: %v", sim)
}
