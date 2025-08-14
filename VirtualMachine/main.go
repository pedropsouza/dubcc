package main

import (
	"bytes"
	"dubcc"
	"dubcc/assembler"
	_ "embed"
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
	"image/png"
	_ "image/png"
	"log"
	"os"
)

var window *app.Window
var register *app.Window
var editor EditorApp
var th *material.Theme
var assembleBtn widget.Clickable
var stepBtn widget.Clickable
var resetBtn widget.Clickable

var memCap dubcc.MachineAddress
var sim dubcc.Sim
var assemblerSingleton assembler.Info
var assemblerInfo assembler.Info

//go:embed logoData.png
var logoData []byte

func main() {
	memCap = dubcc.MachineAddress(1 << 6)
	sim = dubcc.MakeSim(memCap)
	assemblerSingleton = assembler.MakeAssembler()
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
		window = new(app.Window)
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

var logoWidget widget.Image

func init() {
	var err error
	logoData, err = os.ReadFile("logoData.png")
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(logoData))
	if err != nil {
		log.Fatalf("Falha ao decodificar a imagem: %v", err)
	}
	logoWidget = widget.Image{
		Src: paint.NewImageOp(img),
		Fit: widget.Contain,
	}
}
