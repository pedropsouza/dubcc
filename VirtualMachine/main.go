package main

import (
	"dubcc/datatypes"
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"log"
	"os"
	"fmt"
	"path"
)

var editor widget.Editor
var assembleBtn widget.Clickable
var stepBtn widget.Clickable

var memCap datatypes.MachineAddress 
var sim datatypes.Sim

func main() {
	memCap = datatypes.MachineAddress(1 << 6)
	sim = datatypes.MakeSim(memCap)
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
