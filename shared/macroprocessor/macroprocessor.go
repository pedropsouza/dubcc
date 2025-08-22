package macroprocessor

import (
	"dubcc"
	"errors"
	"fmt"
	"log"

	//"slices"
	"strings"
)

type Info struct {
	macros       map[string]*Macro
	macroLevel   int
	macroStack   []MacroMeta
	output       []string
	lineCounter dubcc.MachineAddress
}

func (info *Info) GetOutput() []string {
	return info.output
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

func (info *Info) ProcessLine(rawline string) (err error) {
	line, err := dubcc.ParseAsmLine(rawline)
	switch line.Op {
	case "MACRO":
		info.macroLevel++
		return nil
	case "MEND":
		if info.macroLevel <= 0 {
			return errors.New("End of macro before start.")
		}
		info.macroLevel--
		meta := info.macroStack[len(info.macroStack)-1]
		info.macroStack = info.macroStack[:len(info.macroStack)-1]
		info.macros[meta.name] = &Macro{
			args: meta.args,
			body: meta.body,
			definedAt: dubcc.MachineAddress(info.lineCounter),
		}
		return nil
	}

	if info.macroLevel > 0 {
		return info.handleMacroDef(line)
	} else {
	  expansion, err := info.expandAndRunAllMacros(line)
		if err != nil {
			return errors.New("Error expanding macro: " + err.Error())
		}
		info.output = append(info.output, expansion...)
		return nil
	}
}

func (info *Info) handleMacroDef(line dubcc.InLine) (err error) {
	// first line for this macro?
	if (info.macroLevel - len(info.macroStack)) > 0 {
    fields := strings.Fields(line.Raw)
		if len(fields) == 0 {
			return errors.New("bad macro syntax!")
		}
		info.macroStack = append(info.macroStack, MacroMeta{
			name: fields[0],
			args: fields[1:],
			body: []string{},
		})
		return nil
	}
	
  expansion, err := info.expandAndRunAllMacros(line)
	if err != nil {
		return errors.New("Error expanding macro: " + err.Error())
	}

	info.macroStack[len(info.macroStack)-1].body = append(info.macroStack[len(info.macroStack)-1].body, expansion...)

	return nil
}

func (info *Info) expandAndRunAllMacros(line dubcc.InLine) (out []string, err error) {
	lines := []dubcc.InLine{line}
	for len(lines) > 0 {
		line := lines[0]
		lines = lines[1:]
		macro, mfound := info.macros[line.Op]
		if mfound {
			macro.uses++
			expansion, err := info.expandAndRunMacro(*macro, line)
			if err != nil { return nil, err }
			for _, expanded_line := range expansion {
				reparse, err := dubcc.ParseAsmLine(expanded_line)
				if err != nil { return nil, err }
				lines = append(lines, reparse)
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

func MakeMacroProcessor() Info {
	return Info{
		macros:     make(map[string]*Macro),
		macroLevel: 0,
		output: make([]string,0),
	}
}

