package main

import (
	_ "bufio"
	"bytes"
	_ "dubcc"
	"dubcc/assembler"
	"io"
	"log"
	"os"

	"github.com/k0kubun/pp/v3"
)

func main() {
	var r io.Reader = os.Stdin

	if len(os.Args) == 2 {
		inputFile, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Println(err)
		}
		r = bytes.NewReader(inputFile)
	}

	obj, err := assembler.Read(r)
	if err != nil {
		panic(err.Error())
	}
	pp.Println(obj)
}
