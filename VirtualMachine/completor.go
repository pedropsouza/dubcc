package main

import (
	"dubcc/assembler"
	"gioui.org/io/key"
	"github.com/oligo/gvcode"
	"strings"
	"unicode/utf8"
)

type AsmCompletor struct {
	editor *gvcode.Editor
}

func isSymbolSeperator(ch rune) bool {
	if (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' {
		return false
	}

	return true
}

func (c *AsmCompletor) Trigger() gvcode.Trigger {
	return gvcode.Trigger{
		Characters: []string{},
		KeyBinding: struct {
			Name      key.Name
			Modifiers key.Modifiers
		}{
			Name: "P", Modifiers: key.ModShortcut,
		},
	}
}

func (c *AsmCompletor) Suggest(ctx gvcode.CompletionContext) []gvcode.CompletionCandidate {
	prefix := c.editor.ReadUntil(-1, isSymbolSeperator)
	candidates := make([]gvcode.CompletionCandidate, 0)
	for kw, instruction := range sim.Isa.Instructions { //Instruction
		if strings.Contains(kw, prefix) {
			candidates = append(candidates, gvcode.CompletionCandidate{
				Label: kw,
				TextEdit: gvcode.TextEdit{
					NewText: kw,
					EditRange: gvcode.EditRange{
						Start: gvcode.Position{Runes: ctx.Position.Runes - utf8.RuneCountInString(prefix)},
						End:   gvcode.Position{Runes: ctx.Position.Runes},
					},
				},
				Description: instruction.Name,
				Kind:        "Instruction",
			})
		}
	}

	for kw, register := range sim.Isa.Registers { //Register
		if strings.Contains(kw, prefix) {
			candidates = append(candidates, gvcode.CompletionCandidate{
				Label: kw,
				TextEdit: gvcode.TextEdit{
					NewText: kw,
					EditRange: gvcode.EditRange{
						Start: gvcode.Position{Runes: ctx.Position.Runes - utf8.RuneCountInString(prefix)},
						End:   gvcode.Position{Runes: ctx.Position.Runes},
					},
				},
				Description: register.Name,
				Kind:        "Register",
			})
		}
	}

	for kw := range assembler.Directives() { //Directives
		if strings.Contains(kw, prefix) {
			candidates = append(candidates, gvcode.CompletionCandidate{
				Label: kw,
				TextEdit: gvcode.TextEdit{
					NewText: kw,
					EditRange: gvcode.EditRange{
						Start: gvcode.Position{Runes: ctx.Position.Runes - utf8.RuneCountInString(prefix)},
						End:   gvcode.Position{Runes: ctx.Position.Runes},
					},
				},
				Description: kw,
				Kind:        "Directive",
			})
		}
	}
	/*
		labels := assembler.Info.GetLabel() //Phew, something is really wrong here eh
		for _, label := range labels { //Labels
			if strings.Contains(label, prefix) {
				candidates = append(candidates, gvcode.CompletionCandidate{
					Label: label,
					TextEdit: gvcode.TextEdit{
						NewText: label,
						EditRange: gvcode.EditRange{
							Start: gvcode.Position{Runes: ctx.Position.Runes - utf8.RuneCountInString(prefix)},
							End:   gvcode.Position{Runes: ctx.Position.Runes},
						},
					},
					Description: label,
					Kind:        "Label",
				})
			}
		}
	*/
	return candidates
}
