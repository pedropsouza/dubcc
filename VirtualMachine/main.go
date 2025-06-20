package main

import (
	"dubcc/datatypes"
	"fmt"
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"log"
	"os"
	"path"
)

var editor widget.Editor
var assembleBtn widget.Clickable
var stepBtn widget.Clickable
var resetBtn widget.Clickable

var memCap datatypes.MachineAddress
var sim datatypes.Sim

func main() {
	memCap = datatypes.MachineAddress(1 << 6)
	sim = datatypes.MakeSim(memCap)
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
			editor.SetText(string(code))
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
