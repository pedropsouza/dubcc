package main

import (
	"dubcc"
	"dubcc/assembler"
	"image"
	"image/color"
	"log"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func textLayout(gtx layout.Context, th *material.Theme, title string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
					return FillWithLabel(gtx, th, title, red, 16)
				}),
				layout.Flexed(0.85, func(gtx layout.Context) layout.Dimensions {
					return FillWithText(gtx, th, "Digite aqui...", white)
				}),
			)
		}))
}

func TextWithTable(gtx layout.Context, th *material.Theme, title string, bg color.NRGBA, table *Table, colWeights []float32) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, title, red, 16)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return table.Draw(gtx, th, colWeights)
			})
		}),
	)
}

func FillWithText(gtx layout.Context, th *material.Theme, text string, bg color.NRGBA) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(8)}.Layout,
		),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			ColorBox(gtx, gtx.Constraints.Max, bg)
			return editor.Layout(gtx, th)
		}),
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(8)}.Layout,
		),
	)
}

func FillWithLabel(gtx layout.Context, th *material.Theme, text string, bg color.NRGBA, size unit.Sp) layout.Dimensions {
	ColorBox(gtx, gtx.Constraints.Max, bg)
	label := material.H3(th, text)
	label.Color = yellow
	label.TextSize = size
	label.Font.Weight = font.ExtraBold
	label.Font.Typeface = customFont
	return layout.Center.Layout(gtx, label.Layout)
}

func ColorBox(gtx layout.Context, size image.Point, c color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: c}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

func actionButtonsLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {

	assembleBtnView := material.Button(th, &assembleBtn, "Assemble")
	assembleBtnView.Background = yellow
	assembleBtnView.Color = black
	assembleBtnView.Font.Typeface = customFont

	stepBtnView := material.Button(th, &stepBtn, "Step")
	stepBtnView.Background = yellow
	stepBtnView.Color = black
	stepBtnView.Font.Typeface = customFont

	resetBtnView := material.Button(th, &resetBtn, "Reset")
	resetBtnView.Background = yellow
	resetBtnView.Color = black
	resetBtnView.Font.Typeface = customFont

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceSides,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if assembleBtn.Clicked(gtx) {
				CompileCode()
			}
			return assembleBtnView.Layout(gtx)
		}),
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(8)}.Layout,
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if sim.State != dubcc.SimStateHalt {
				if stepBtn.Clicked(gtx) {
					StepSimulation()
				}
				return stepBtnView.Layout(gtx)
			} else {
				if resetBtn.Clicked(gtx) {
					log.Printf("reset!")
					sim.State = dubcc.SimStateRun
					sim.Registers = dubcc.StartupRegisters(&sim.Isa, dubcc.MachineAddress(len(sim.Mem.Work)))
					startAddressMachineWord := dubcc.MachineWord(assemblerInfo.StartAddress)
					sim.SetRegister(dubcc.RegPC, startAddressMachineWord)
				}
				return resetBtnView.Layout(gtx)
			}
		}),
	)
}

func CompileCode() {
	sim.Registers = dubcc.StartupRegisters(&sim.Isa, dubcc.MachineAddress(len(sim.Mem.Work)))
	assemblerInfo = assembler.MakeAssembler()
	for _, line := range strings.Split(editor.state.Text(), "\n") {
		assemblerInfo.FirstPassString(line)
	}
	assemblerInfo.SecondPass()
	mem := assemblerInfo.GetOutput()
	if len(mem) > len(sim.Mem.Work) {
		panic("program's too big")
	}
	startAddressMachineWord := dubcc.MachineWord(assemblerInfo.StartAddress)
	sim.SetRegister(dubcc.RegPC, startAddressMachineWord) // Altera o valor do PC pro valor indicado na diretiva "start"
	startAddressInt := int(assemblerInfo.StartAddress)
	for idx, val := range mem {
		sim.Mem.Work[startAddressInt+idx] = val
	}
	sim.State = dubcc.SimStatePause
}

func StepSimulation() {
	pc := sim.GetRegister(dubcc.RegPC)
	instWord := sim.Mem.Work[pc]
	sim.SetRegister(dubcc.RegRI, instWord)
	inst, ifound := sim.InstructionFromWord(instWord)
	if !ifound {
		log.Printf("invalid instruction %x (%d)\n", instWord, instWord)
	} else {
		handler, hfound := sim.Handlers[inst.Repr]
		if !hfound {
			log.Printf("couldn't handle instruction %v\n", inst)
		}
		instPos := dubcc.MachineAddress(pc)
		argsTerm := instPos + 1 + dubcc.MachineAddress(inst.NumArgs)
		// set pc before calling the handler
		// that way branching works
		nextPc := (pc + dubcc.MachineWord(1+inst.NumArgs)) % dubcc.MachineWord(len(sim.Mem.Work))
		sim.MapRegister(
			dubcc.RegPC,
			func(pc dubcc.MachineWord) dubcc.MachineWord {
				return nextPc
			},
		)
		if nextPc < dubcc.MachineWord(instPos) {
			log.Printf("pc wrapped around! halt.")
			sim.State = dubcc.SimStateHalt
			return
		}
		args := sim.Mem.Work[instPos:argsTerm]
		log.Printf("Executing %s with %v", inst.Name, args)
		handler(&sim, args)
		if sim.State != dubcc.SimStateHalt {
			sim.State = dubcc.SimStatePause
		}
	}
}
