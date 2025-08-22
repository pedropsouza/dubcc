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
	macroStack   []MacroFrame
	output       []string
	lineCounter dubcc.MachineAddress
}

func (info *Info) GetOutput() []string {
	return info.output
}

type Macros struct {
	args      []string
	body      []string
	definedAt dubcc.MachineAddress //Totalmente opcional, uso futuro para mensagens de erro
}

type MacroFrame struct {
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

var EmptyLineErr = errors.New("empty line")

func (info *Info) ProcessLine(rawline string) (err error) {
	line, err := dubcc.ParseAsmLine(rawline)
	if line.Op == "MACRO" {
		info.macroLevel++
		return nil
	}
	if info.macroLevel > 0 {
		return info.handleMacroDef(line)
	}

	expansion := line.Raw
	macro, mfound := info.macros[line.Op]
	if mfound { //This shit has to be a macro, right?
		expansion, err = info.expandAndRunMacro(macro, line)
		if err != nil { return err }
	}
	if line.Op == "MEND" { //De preferência, deixar como último teste
		return errors.New("End of macro before start.")
	}

	info.output = append(info.output, expansion)
	return nil
}

func (info *Info) handleMacroDef(line dubcc.InLine) (err error) {
	if len(info.macroStack) < info.macroLevel { //Se for a primeira linha...
		for len(info.macroStack) < info.macroLevel {
			info.macroStack = append(info.macroStack, MacroFrame{})
		}
		info.macroStack[info.macroLevel-1] = MacroFrame{
			name: line.Op,
			args: line.Args,
			body: []string{},
		}
		return nil
	}

	if line.Op == "MEND" { //Aqui, toda macro foi lida e o MEND vai fechar a macro
		info.macroLevel--
		frame := info.macroStack[info.macroLevel]
		macro := Macros{
			args:      frame.args,
			body:      frame.body,
			definedAt: info.lineCounter,
		}

		info.macroStack = slices.Delete(info.macroStack, info.macroLevel, info.macroLevel+1)
		info.macros[frame.name] = macro
		return nil
	}
	frame := &info.macroStack[info.macroLevel-1]
	frame.body = append(frame.body, line.Raw)
	return nil
}

func (info *Info) expandAndRunMacro(macro Macros, line dubcc.InLine) (string, error) {
	macro_err_str := "macro error!"
	if len(line.Args) != len(macro.args) {
		return macro_err_str, errors.New("number of arguments doesn't match")
	}

	substitutions := make(map[string]string)
	for i, formal := range macro.args {
		substitutions[formal] = line.Args[i]
	}

	var macro_expansion = ""
	for _, raw := range macro.body {
		words := strings.Split(raw, " ")
		for i, word := range words {
			wdata, wfound := substitutions[word]
			if wfound {
				words[i] = wdata
			}
		}
		expanded := strings.Join(words, " ")

		macro_expansion = macro_expansion + expanded
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

