package main

import (
	"dubcc"
	"bufio"
	"bytes"
	linker "dubcc/linker"
	assembler "dubcc/assembler"
	"github.com/k0kubun/pp/v3"
	"log"
	"os"
)

const (
	Relocator = linker.Relocator
	Absolute = linker.Absolute
)

type MachineAddress =  dubcc.MachineAddress
type ObjectFile = assembler.ObjectFile
type SourceFile = assembler.SourceFile

var linkerMode LinkerMode
var loadAddress MachineAddress
var objects []*ObjectFile

func main() {
	if len(os.Args) >= 2 {
		sourceAlreadyRead := false
		linkerMode = Relocator
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			switch arg {
			case "-a", "--absolute":
				if len(os.Args) == i+1 {
					log.Fatal("error: --absolute requires i+1 arg load address")
				} else {
					linkerMode = Absolute
					loadAddress = ParseNum(os.Args[i+1])
				} 
			default:
				code, err := os.ReadFile(arg)
				if err != nil {
					log.Printf("%v\n", err)
					continue
				}
				file := SourceFile{
					Name: string(arg),
					Data: string(code),
					Object: nil,
				}
				files = append(files, file)
				if !sourceAlreadyRead {
					editor.state.SetText(files[0].Data)
					sourceAlreadyRead = true
				}
	  	}
		} 
	} else {
		log.Fatal("error: linker needs >=2 input files")
	}

	switch linkerMode {
	case Relocator:
		linkerSingleton := linker.MakeRelocatorLinker()
	case Absolute:
		linkerSingleton := linker.MakeAbsoluteLinker(loadAddress)
	}

	for i := range files {
		objects = append(objects, files[i].Object)
	}

	executable, err := linkerSingleton.GenerateExecutable(objects)

	if err != nil {
		log.Printf("error: could not generate an executable\n%s\n", err.Error())
	}
}
