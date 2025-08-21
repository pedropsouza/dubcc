package macroprocessor

import (
	"dubcc"
	"errors"
	"log"
	"slices"
	"strings"
)

type Info struct {
	macros       map[string]Macros
	macroLevel   int
	//macroStack   []MacroFrame
	output       []string
	lineCounter dubcc.MachineAddress
}

func (info *Info) GetOutput() []string {
	return info.output
}

type Macro struct {
	args      []string
	body      []string
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
  if line.Op == "MACRO" {
		info.macroLevel++
		return nil
	} else if line.Op == "MEND" {
		if info.macroLevel <= 0 {
			return errors.New("End of macro before start.")
		}
		info.macroLevel--
	}

	expansion, err := info.expandAndRunAllMacros(line)
	if err != nil {
		return errors.New("Error expanding macro: " + err.Error())
	}
	info.output = append(info.output, expansion...)
	return nil
}

func (info *Info) handleMacroDef(acc *Macro, line dubcc.InLine) (err error) {
	if len(acc.body) == 0 {
    fields := strings.Fields(line.Raw)
		if len(fields) == 0 {
			return errors.New("bad macro syntax!")
		}
		acc = MacroMeta{
			name: fields[0],
			args: fields[1:],
			body: []string{},
		}
		return nil
	}
}

func (info *Info) expandAndRunAllMacros(line dubcc.InLine) (out []string, err error) {
	lines := []dubcc.InLine{line}
	for len(lines) > 0 {
		line := lines[0]
		lines = lines[1:]
		macro, mfound := info.macros[line.Op]
		if mfound {
			expansion, err := info.expandAndRunMacro(macro, line)
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

func (info *Info) expandAndRunMacro(macro Macros, line dubcc.InLine) ([]string, error) {
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

func MakeMacroProcessor() Info {
	return Info{
		macros:     make(map[string]Macros),
		macroLevel: 0,
		macroStack: make([]MacroFrame,0),
		output: make([]string,0),
	}
}

