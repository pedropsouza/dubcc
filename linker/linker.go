package main

import (
	"dubcc"
	"dubcc/linker"
	"dubcc/assembler"
	"log"
	"os"
	"strings"
	"bytes"
	"path/filepath"
)

const (
	Relocator = linker.Relocator
	Absolute = linker.Absolute
)

type (
	Linker	= linker.Linker
	LinkerMode = linker.LinkerMode
	MachineAddress =  dubcc.MachineAddress
	ObjectFile = assembler.ObjectFile
	SourceFile = assembler.SourceFile
)

var files	[]SourceFile
var objects []*ObjectFile
var linkerMode LinkerMode
var loadAddress MachineAddress

func main() {
	if len(os.Args) >= 2 {
		linkerMode = Relocator
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			switch arg {
			case "-a", "--absolute":
				if len(os.Args) == i+1 {
					log.Fatal("error: --absolute requires i+1 arg load address")
				} else {
					linkerMode = Absolute
					var err error
					loadAddress, err = assembler.ParseNum(os.Args[i+1])
					if err != nil {
						log.Fatal("error: failed to parse num")
					}
				} 
			default:
				code, err := os.ReadFile(arg)
				if err != nil {
					log.Printf("%v\n", err)
					continue
				}
				r := bytes.NewReader(code)
				obj, err := assembler.Read(r)
				if err != nil {
					panic(err.Error())
				}
				file := SourceFile{
					Name: string(arg),
					Data: "",
					Object: obj,
				}
				files = append(files, file)
	  	}
		} 
	} else {
		log.Fatal("error: linker needs >=2 input files")
	}

	var linkerSingleton *Linker
	switch linkerMode {
	case Relocator:
		linkerSingleton = linker.MakeRelocatorLinker()
	case Absolute:
		linkerSingleton = linker.MakeAbsoluteLinker(loadAddress)
	}

	for i := range files {
		objects = append(objects, files[i].Object)
	}

	executable, err := linkerSingleton.GenerateExecutable(objects)
	if err != nil {
		log.Printf("error: could not generate an executable\n%s\n", err.Error())
	}

	path := files[0].Name
	base := filepath.Base(path)
	if dot := strings.LastIndex(base, "."); dot != -1 {
		base = base[:dot]
	}
	objFilename := base + ".hpx"
	if err := assembler.SaveCompleteObjectFile(executable, objFilename); err != nil {
		log.Printf("warning: could not save %s: %v", objFilename, err)
	}
}
