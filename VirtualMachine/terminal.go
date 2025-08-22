package main

import (
	"dubcc"
	"strings"
	"sync"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Terminal struct {
	editorTerminal widget.Editor
	scrollArea     widget.List
	lines          []string
	mu             sync.RWMutex
	theme          *material.Theme
	inputChan      chan string
	waiting        bool
}

func (t *Terminal) Clear() {
	t.lines = make([]string, 1)
}

func NewTerminal(theme *material.Theme) *Terminal {
	t := &Terminal{
		theme:     theme,
		lines:     make([]string, 1),
		inputChan: make(chan string, 1),
	}

	t.editorTerminal.SingleLine = true
	t.editorTerminal.Submit = true
	t.scrollArea.Axis = layout.Vertical
	return t
}

func (t *Terminal) Write(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	inLines := strings.Split(text, "\n")
	targetLine := &t.lines[len(t.lines)-1]
	*targetLine += inLines[0]
	for _, xtraLine := range inLines[1:] {
		t.lines = append(t.lines, xtraLine)
	}

	t.scrollArea.Position.Offset = 1e6
}

func (t *Terminal) Read() string {
	t.waiting = true
	input := <-t.inputChan
	t.waiting = false
	return input
}

func (t *Terminal) layoutOutput(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			t.mu.RLock()
			lineCount := len(t.lines)
			t.mu.RUnlock()

			return t.scrollArea.Layout(gtx, lineCount, func(gtx layout.Context, i int) layout.Dimensions {
				t.mu.RLock()
				if i >= len(t.lines) {
					t.mu.RUnlock()
					return layout.Dimensions{}
				}
				line := t.lines[i]
				t.mu.RUnlock()

				label := material.Label(t.theme, unit.Sp(16), line)
				label.Color = white
				return label.Layout(gtx)
			})
		})
}

func (t *Terminal) layoutInput(gtx layout.Context) layout.Dimensions {
	for {
		event, ok := t.editorTerminal.Update(gtx)
		if !ok {
			break
		}

		if _, ok := event.(widget.SubmitEvent); ok {
			input := strings.TrimSpace(t.editorTerminal.Text())
			if input != "" {
				t.Write("> " + input)

				sim.TxInWord(dubcc.MachineWord([]byte(input)[0]))

				select {
				case t.inputChan <- input:
				default:
				}

				t.editorTerminal.SetText("")
			}
		}
	}

	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			prompt := "> "
			if t.waiting {
				prompt = "? "
			}

			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Label(t.theme, unit.Sp(12), prompt)
					label.Color = yellow
					return label.Layout(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					editorTerminal := material.Editor(t.theme, &t.editorTerminal, "")
					editorTerminal.Color = yellow
					return editorTerminal.Layout(gtx)
				}),
			)
		})
}

func LayoutGeral(gtx layout.Context, terminal *Terminal) layout.Dimensions {
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, black, rect.Op())
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return terminal.layoutOutput(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return terminal.layoutInput(gtx)
		}),
	)
}
