package macroprocessor

import (
	"dubcc"
	_ "dubcc/assembler"
	"errors"
	"fmt"
	"log"

	//"slices"
	"strings"
)

const (
	GND = 0
	META = 1
	BODY = 2
	// all subsequent states are clones of body
)

type Info struct {
	macros       map[string]*Macro
	state        int
	currentDef   *MacroMeta
	lineCount    int
}

type Macro struct {
	args      []string
	body      []string
	uses      int
	definedAt dubcc.MachineAddress //Totalmente opcional, uso futuro para mensagens de erro
}

type MacroMeta struct {
	name string
  args []string
	body []string
}

func BoolToInt(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

func (info *Info) ProcessLine(rawline string) (output []string, err error) {
	line, err := dubcc.ParseAsmLine(rawline)
	if err != nil { return nil, err }
	if line.Op == "MACRO" {
		info.state++
	}

	if info.state != GND {
		log.Printf("line \"%s\" with macro level %d", rawline, info.state)
		err := info.handleMacroDef(line)
		if line.Op == "MEND" {
			info.state--
			if info.state == META {
				info.state--
			}
		}
		return nil,err
	}

	log.Printf("line \"%s\" with macro level %d", rawline, info.state)
	expansion, err := info.expandAndRunAllMacros(line)
	if err != nil {
		return nil,err
	}
	output = append(output, expansion...)
	return output, nil
}

func (info *Info) handleMacroDef(line dubcc.InLine) (err error) {
	// skip MACRO line
	if info.state == META { info.state++; return nil }
	// first body line for this macro?
	if info.currentDef == nil && info.state == BODY {
    fields := strings.Fields(line.Raw)
		if len(fields) == 0 {
			return errors.New("bad macro syntax!")
		}
		info.currentDef = &MacroMeta{
			name: fields[0],
			args: fields[1:],
			body: []string{},
		}
		return nil
	}
	
	if line.Op == "MEND" && info.state == BODY {
		meta := info.currentDef
		info.macros[meta.name] = &Macro{
			args: meta.args,
			body: meta.body,
			definedAt: dubcc.MachineAddress(info.lineCount),
		}
		info.currentDef = nil
		return nil
	}

	if info.currentDef != nil {
		info.currentDef.body = append(info.currentDef.body, line.Raw)
	}

	return nil
}

func (info *Info) expandAndRunAllMacros(line dubcc.InLine) (out []string, err error) {
	lines := []dubcc.InLine{line}
	for len(lines) > 0 {
		line := lines[0]
		lines = lines[1:]
		macro, mfound := info.macros[line.Op]
		if mfound {
			expansion, err := info.expandAndRunMacro(*macro, line)
			if err != nil { return nil, err }
			for _, expanded_line := range expansion {
				innerLines, err := info.ProcessLine(expanded_line)
				if err != nil { return nil, err }
				reparse := []dubcc.InLine{}
				for _, line := range innerLines {
					reparsedLine, err := dubcc.ParseAsmLine(line)
					if err != nil { return nil, err }
					reparse = append(reparse, reparsedLine)
				}
				lines = append(lines, reparse...)
			}
		} else {
			out = append(out, line.Raw)
		}
	}
	return out, err
}

func (info *Info) expandAndRunMacro(macro Macro, line dubcc.InLine) ([]string, error) {
	macro_err_str := []string{"macro error!"}
	if len(line.Args) != len(macro.args) {
		return macro_err_str, errors.New("number of arguments doesn't match")
	}

	substitutions := make(map[string]string)
	for i, formal := range macro.args {
		substitutions[formal] = line.Args[i]
	}

	var macro_expansion = []string{}
	if line.Label != "" {
		macro_expansion = append(macro_expansion, line.Label + ":")
	}
	for _, raw := range macro.body {
		words := strings.Split(raw, " ")
		for i, word := range words {
			wdata, wfound := substitutions[word]
			if wfound {
				words[i] = wdata
			}
		}
		expanded := strings.Join(words, " ")

		macro_expansion = append(macro_expansion, expanded)
	}
	log.Print(macro_expansion)
	return macro_expansion, nil
}

func (info *Info) MacroReport() string {
	return fmt.Sprintf("Macro report:\n\t%d macros", len(info.macros))
}

func MakeMacroProcessor(lineNum int) Info {
	return Info{
		macros: make(map[string]*Macro),
		state: GND,
		lineCount: lineNum,
	}
}

