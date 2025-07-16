package main

import (
	"dubcc"
	"dubcc/assembler"
	"fmt"
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gvcode"
	"github.com/oligo/gvcode/addons/completion"
	wg "github.com/oligo/gvcode/widget"
	"log"
	"os"
)

var editor EditorApp
var th *material.Theme
var assembleBtn widget.Clickable
var stepBtn widget.Clickable
var resetBtn widget.Clickable

var memCap dubcc.MachineAddress
var sim dubcc.Sim
var assemblerInfo assembler.Info

func main() {
	memCap = dubcc.MachineAddress(1 << 6)
	sim = dubcc.MakeSim(memCap)
	fmt.Println(sim)
	InitTables(&sim)
	editor = EditorApp{}
	th = material.NewTheme()
	editor.state = wg.NewEditor(th)
	gvcode.SetDebug(false)

	if len(os.Args) > 1 {
		code, err := os.ReadFile(os.Args[1])
		if err == nil {
			editor.state.SetText(string(code))
		} else {
			log.Printf("%v\n", err)
		}
	}

	go func() {
		window := new(app.Window)
		window.Option(app.Title("Dobam Unka Bee Compiler Collection"))

		if err := run(window); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	var ops op.Ops

	customScheme := createCustomColorScheme(th)

	// Setting up auto-completion.
	cm := &completion.DefaultCompletion{Editor: editor.state}

	// set popup widget to let user navigate the candidates.
	popup := completion.NewCompletionPopup(editor.state, cm)
	popup.Theme = th
	popup.TextSize = unit.Sp(12)

	cm.AddCompletor(&AsmCompletor{editor: editor.state}, popup)

	editor.state.WithOptions(
		gvcode.WrapLine(false),
		gvcode.WithLineNumber(true),
		gvcode.WithAutoCompletion(cm),
		gvcode.WithColorScheme(customScheme),
	)

	tokens := HighlightTextByPattern(editor.state.Text())
	editor.state.SetSyntaxTokens(tokens...)
	editor.state.SetText(editor.state.Text())

	/*editor.state.AddDecorations(
		decoration.Decoration{Source: "test", Start: 5, End: 150, Background: &decoration.Background{Color: highlightColor}},
		decoration.Decoration{Source: "test", Start: 100, End: 200, Background: &decoration.Background{Color: highlightColor2}},
		decoration.Decoration{Source: "test", Start: 100, End: 200, Squiggle: &decoration.Squiggle{Color: highlightColor3}},
		decoration.Decoration{Source: "test", Start: 250, End: 400, Strikethrough: &decoration.Strikethrough{Color: highlightColor3}},
	)*/

	for {
		event := window.Event()
		switch e := event.(type) {
		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			paint.ColorOp{Color: red}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)

			mainLayout(gtx, th)

			e.Frame(gtx.Ops)
		}
	}
}
