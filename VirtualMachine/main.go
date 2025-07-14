package main

import (
	"dubcc"
	"dubcc/assembler"
	"fmt"
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"log"
	"os"
	"path"
	"github.com/oligo/gvcode"
	"github.com/oligo/gvcode/addons/completion"
	gvcolor "github.com/oligo/gvcode/color"
	"github.com/oligo/gvcode/textstyle/decoration"
	"github.com/oligo/gvcode/textstyle/syntax"
	wg "github.com/oligo/gvcode/widget"
)

var editor EditorApp
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
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(execPath)
	pathEnv := os.Getenv("PATH")
	assemblerDir := path.Join(path.Dir(path.Dir(execPath)), "assembler")
	os.Setenv("PATH", pathEnv+":"+assemblerDir)
	log.Print(os.Getenv("PATH"))

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
		window.Option(app.Title("dobam unka bee compiler collection (speed racer)"))

		if err := run(window); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	th := material.NewTheme()
	var ops op.Ops
	editor.SingleLine = false

	editor := EditorApp{
		th: th,
	}
	editorApp.state = wg.NewEditor(th)
	gvcode.SetDebug(false)
	
	// color scheme
	colorScheme := syntax.ColorScheme{}
	colorScheme.Foreground = gvcolor.MakeColor(th.Fg)
	colorScheme.SelectColor = gvcolor.MakeColor(th.ContrastBg).MulAlpha(0x60)
	colorScheme.LineColor = gvcolor.MakeColor(th.ContrastBg).MulAlpha(0x30)
	colorScheme.LineNumberColor = gvcolor.MakeColor(th.ContrastBg).MulAlpha(0xb6)
	keywordColor, _ := gvcolor.Hex2Color("#AF00DB")
	colorScheme.AddStyle("keyword", syntax.Underline, keywordColor, gvcolor.Color{})

	editorApp.state.WithOptions(
		gvcode.WrapLine(false),
		gvcode.WithLineNumber(true),
		gvcode.WithAutoCompletion(cm),
		gvcode.WithColorScheme(colorScheme),
	)

	tokens := HightlightTextByPattern(editorApp.state.Text(), syntaxPattern)
	editorApp.state.SetSyntaxTokens(tokens...)

	highlightColor, _ := gvcolor.Hex2Color("#e74c3c50")
	highlightColor2, _ := gvcolor.Hex2Color("#f1c40f50")
	highlightColor3, _ := gvcolor.Hex2Color("#e74c3c")

	editorApp.state.AddDecorations(
		decoration.Decoration{Source: "test", Start: 5, End: 150, Background: &decoration.Background{Color: highlightColor}},
		decoration.Decoration{Source: "test", Start: 100, End: 200, Background: &decoration.Background{Color: highlightColor2}},
		decoration.Decoration{Source: "test", Start: 100, End: 200, Squiggle: &decoration.Squiggle{Color: highlightColor3}},
		decoration.Decoration{Source: "test", Start: 250, End: 400, Strikethrough: &decoration.Strikethrough{Color: highlightColor3}},
	)

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
