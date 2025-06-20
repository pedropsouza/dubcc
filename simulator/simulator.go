package main

import (
	"fmt"
	"io"
	"bufio"
	"log"
	"os"
	"os/signal"
	"time"
	"dubcc"
	"github.com/k0kubun/pp/v3"
)

func main() {
	memCap := dubcc.MachineAddress(1<<5)
	sim := dubcc.MakeSim(memCap) // 32b for now
	log.Printf("loaded %d", memCap)
	reader := bufio.NewReader(os.Stdin)

	{ // install interrupt handler
		c := make (chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for _ = range c {
				sim.State = dubcc.SimStateHalt
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
			v := dubcc.MachineWord(buf[0] << 8 + buf[1])
			fmt.Fprintf(os.Stderr, "got word %x (%d) out of %v\n", v, v, buf)
			sim.Mem.Work[mempos] = v
		}
	}

	sim.State = dubcc.SimStateRun
	for {
		if sim.State != dubcc.SimStateRun {
			break
		}
		pc := sim.Isa.Registers["PC"]
		ri := sim.Isa.Registers["RI"]
		//re := sim.Isa.Registers["RE"]
		ri.Content = sim.Mem.Work[pc.Content]
		inst, ifound := sim.Isa.InstructionFromWord(ri.Content)
		if !ifound {
			fmt.Fprintf(os.Stderr,
				"invalid instruction %x (%d)\n", ri.Content, ri.Content,
			)
		}
		handler, hfound := sim.Handlers[inst.Repr]
		if !hfound {
			fmt.Fprintf(os.Stderr, "couldn't handle instruction %v\n", inst)
		}
		instPos := dubcc.MachineAddress(pc.Content)
		argsTerm := instPos + 1 + dubcc.MachineAddress(inst.NumArgs)
		// set pc before calling the handler
		// that way branching works
		pc.Content += dubcc.MachineWord(1 + inst.NumArgs)
		args := sim.Mem.Work[instPos:argsTerm]
		log.Printf("Executing %s with %v", inst.Name, args)
		handler(&sim, args)
		time.Sleep(100 * time.Millisecond)
	}
	pp.Printf("Simulation state: %v", sim)
}
