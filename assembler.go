package main

import (
	"fmt"
	"errors"
	"strconv"
)

type Info struct {
	isa ISA
	symbols map[string]uint64
	output []byte
}

type ISA struct {
	instructions map[string]Instruction
}

type Instruction struct {
	numArgs byte
	repr byte
}

type Line struct {
	reprs []Repr
}

type ReprKind uint8
const (
	ReprRaw ReprKind = iota
	ReprPartial
	ReprComplete
)

type Repr {
	tag ReprKind
	input string
	symbol string
	out byte
}

func (info *Info) firstPass(line) (repr []IRepr, err error) {
	isa := info.isa
	idata := isa.instructions[line.op]
	r := make([]byte, 1+idata.numArgs)
	if idata.numArgs != len(line.args) {
		return nil, errors.New("number of arguments doesn't match")
	}
	r.append(idata.repr)
	
	for arg := range line.args {
		// try constant interpretation
		num, err := parseNum(arg)
		if err != nil {
			r.append(byte(num)) // will overflow, panic maybe?
			continue
		}
		// aight it ain't a number
		// check symbol table
		lookup := 
	}
	
	return r, nil
}

func (in *Repr) parseNum() (num uint64, err error) {
	if in.tag != ReprRaw { return nil, errors.New("repr is already parsed") }
	b2 := regexp.MustCompile("^0b[0-1]+$")
	b8 := regexp.MustCompile("^0o[0-7]+$")
	b10 := regexp.MustCompile("^[0-9]+$")
	b16 := regexp.MustCompile("^0x[0-9]+$")
	switch {
	case b2.MatchString(in): return strconv.ParseInt(in, 2, 64)
	case b8.MatchString(in): return strconv.ParseInt(in, 8, 64)
	case b10.MatchString(in): return strconv.ParseInt(in, 10, 64)
	case b16.MatchString(in): return strconv.ParseInt(in, 16, 64)
	default: return nil, errors.New("string is not a number")
	}
}

func main() {
	fmt.Println("hello world")
}
