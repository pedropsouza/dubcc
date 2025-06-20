package main

import (
	"bufio"
	"os"
	"log"
	"bytes"
	assembler "dubcc/assembler"
	"github.com/k0kubun/pp/v3"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	if len(os.Args) == 2 {
		inputFile, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Println(err)
		}
		r := bytes.NewReader(inputFile)
		scanner = bufio.NewScanner(r)
	}

	info := assembler.MakeAssembler()
	for {
		if !scanner.Scan() { break }
		line := scanner.Text()
		outLine, err := info.FirstPassString(line)
		pp.Fprintf(os.Stderr, "processing %v... ", line)
		if err != nil {
			if err != assembler.EmptyLineErr {
				log.Println(err)
			}
			continue
		}
		if err != nil {
			log.Println(err)
		}
		pp.Fprintf(os.Stderr, "%v\n", outLine)
	}

	symbols := info.SecondPass()
	pp.Fprintf(os.Stderr, "Symbols: %v\n", symbols)

	{ // write binary
		writer := bufio.NewWriter(os.Stdout)
		for _, u16 := range info.GetOutput() {
			high := byte((u16 >> 8) & 0xff)
			low := byte((u16 >> 0) & 0xff)
			writer.WriteByte(high)
			writer.WriteByte(low)
		}
		writer.Flush()
	}
}
