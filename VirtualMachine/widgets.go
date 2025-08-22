package main

import (
	"dubcc"
	"dubcc/assembler"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (mb *MenuBar) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	bar := layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mb.fileBtn.Clicked(gtx) {
				mb.showFileMenu = !mb.showFileMenu
			}
			btn := material.Button(th, &mb.fileBtn, "File")
			btn.Background = yellow
			btn.Color = black
			btn.Font.Typeface = customFont
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, &mb.editBtn, "Edit")
			btn.Background = yellow
			btn.Color = black
			btn.Font.Typeface = customFont
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mb.helpBtn.Clicked(gtx) {
				mb.showHelpMenu = !mb.showHelpMenu
			}
			btn := material.Button(th, &mb.helpBtn, "Help")
			btn.Background = yellow
			btn.Color = black
			btn.Font.Typeface = customFont
			return btn.Layout(gtx)
		}),
	)
	layout.Stacked(func(gtx layout.Context) layout.Dimensions {
		if !mb.showFileMenu {
			return layout.Dimensions{}
		}
		op.Offset(image.Pt(0, gtx.Dp(unit.Dp(30)))).Add(gtx.Ops)
		return mb.renderFileMenu(gtx, th)
	})
	return bar
}
func (mb *MenuBar) renderFileMenu(gtx layout.Context, th *material.Theme) layout.Dimensions {
	size := image.Pt(gtx.Dp(unit.Dp(90)), gtx.Dp(unit.Dp(120)))
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: yellow}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := TextButton(th, &mb.openBtn, "Open…")
			if mb.openBtn.Clicked(gtx) {
				showExplorer = true
				mb.showFileMenu = false
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := TextButton(th, &mb.saveBtn, "Save")
			if mb.saveBtn.Clicked(gtx) {
				mb.showFileMenu = false
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := TextButton(th, &mb.saveAsBtn, "Save As")
			if mb.saveAsBtn.Clicked(gtx) {
				showSaveDialog = true
				mb.showFileMenu = false
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := TextButton(th, &mb.exitBtn, "Exit")
			if mb.exitBtn.Clicked(gtx) {
				mb.showFileMenu = false
				os.Exit(0)
			}
			return btn.Layout(gtx)
		}),
	)
}

func renderSaveAsDialog(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if !showSaveDialog {
		return layout.Dimensions{}
	}

	size := gtx.Constraints.Max
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 180}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	cardW := int(float32(size.X) * 0.9)
	cardH := int(float32(size.Y) * 0.9)
	offX := (size.X - cardW) / 2
	offY := (size.Y - cardH) / 2
	defer op.Offset(image.Pt(offX, offY)).Push(gtx.Ops).Pop()

	gtx2 := gtx
	gtx2.Constraints = layout.Constraints{
		Min: image.Pt(cardW, cardH),
		Max: image.Pt(cardW, cardH),
	}

	radius := gtx2.Dp(unit.Dp(12))
	defer clip.UniformRRect(image.Rectangle(clip.Rect{Max: image.Pt(cardW, cardH)}), radius).Push(gtx2.Ops).Pop()

	paint.ColorOp{Color: white}.Add(gtx2.Ops)
	paint.PaintOp{}.Add(gtx2.Ops)

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx2, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H5(th, "Save As")
				title.Color = black
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return saveExplorer.Layout(gtx, th)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(th, "Nome do arquivo:")
						lbl.Color = black
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								ed := material.Editor(th, &filenameEditor, "exemplo.asm")
								ed.Color = black
								return ed.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if !strings.HasSuffix(filenameEditor.Text(), ".asm") {
									text := strings.TrimSuffix(filenameEditor.Text(), ".asm")
									if lastDot := strings.LastIndex(text, "."); lastDot != -1 {
										text = text[:lastDot]
									}
								}
								return layout.Dimensions{}
							}),
						)
					}),
				)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if saveCancelBtn.Clicked(gtx) {
							showSaveDialog = false
							filenameEditor.SetText("")
						}
						btn := material.Button(th, &saveCancelBtn, "Cancel")
						btn.Background = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
						btn.Color = white
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if saveConfirmBtn.Clicked(gtx) {
							performSaveAs()
						}
						btn := material.Button(th, &saveConfirmBtn, "Save")
						btn.Background = yellow
						btn.Color = black
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
}

func performSaveAs() {
	filename := strings.TrimSpace(filenameEditor.Text())
	if filename == "" {
		log.Printf("Name of file must contain something.")
		return
	}

	// Garantir extensão .asm
	if !strings.HasSuffix(filename, ".asm") {
		if lastDot := strings.LastIndex(filename, "."); lastDot != -1 {
			filename = filename[:lastDot]
		}
		filename += ".asm"
	}

	currentDir := saveExplorer.current
	fullPath := filepath.Join(currentDir, filename)

	content := editor.state.Text()
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		log.Printf("Error: couldn't save file: %v", err)
	} else {
		log.Printf("File saved succesfully: %s", fullPath)
		showSaveDialog = false
		filenameEditor.SetText("")
	}
}

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

func TextButton(th *material.Theme, btn *widget.Clickable, label string) material.ButtonStyle {
	b := material.Button(th, btn, label)
	b.Background = color.NRGBA{}
	b.Color = black
	b.Inset = layout.UniformInset(unit.Dp(4))
	b.CornerRadius = unit.Dp(0)
	return b
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
				WipeMemory()
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
					WipeMemory()
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
	for idx, val := range mem {
		sim.Mem.Work[idx] = val
	}
	sim.State = dubcc.SimStatePause
}
func WipeMemory() {
	memCap := len(sim.Mem.Work)
	for i := range memCap {
		sim.Mem.Work[i] = 0
	}
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
