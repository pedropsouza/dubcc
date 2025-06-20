package main

import (
	"bytes"
	"dubcc/datatypes"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"io"
	"log"
	"os/exec"
	"strings"
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
		layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, title, red, 16)
		}),
		layout.Flexed(0.85, func(gtx layout.Context) layout.Dimensions {
			ColorBox(gtx, gtx.Constraints.Max, bg)
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
			return material.Editor(th, &editor, text).Layout(gtx)
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

	return layout.Flex{
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceAround, // Distribui o espaço ao redor
		Alignment: layout.Middle,      // Centraliza verticalmente
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if assembleBtn.Clicked(gtx) {
						go func() {
							cmd := exec.Command("assembler")
							if cmd.Err != nil {
								log.Print(cmd.Err)
								return
							}

							cmd.Stdin = strings.NewReader(editor.Text())
							stdout, outerr := cmd.StdoutPipe()
							if outerr != nil {
								log.Print(outerr)
							}

							go func() {
								err := cmd.Start()
								if err != nil {
									log.Print(err)
								}
								err = cmd.Wait()
								if err != nil {
									log.Print(err)
								}
							}()

							data, rerr := io.ReadAll(stdout)
							if rerr != nil {
								log.Print(rerr)
								return
							}
							log.Print(data)
							reader := bytes.NewReader(data)
							{ // read bin to memory
								buf := make([]byte, 2) // read one words worth at a time
								done := false
								for memPos := range sim.Mem.Work {
									if done {
										break
									}
									buf[0] = 0
									buf[1] = 0
									for idx := range buf {
										readb, err := reader.ReadByte()
										if err != nil {
											if err == io.EOF {
												done = true
												break
											}
											log.Fatal("error reading stdin: %v", err)
										}
										buf[idx] = readb
									}
									v := datatypes.MachineWord(buf[0]<<8 + buf[1])
									log.Printf("got word %x (%d) out of %v\n", v, v, buf)
									sim.Mem.Work[memPos] = v
								}
							}
						}()
					}
					return assembleBtnView.Layout(gtx)
				}),

				layout.Rigid(
					layout.Spacer{Width: unit.Dp(16)}.Layout,
				),

				// stepBtn
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					if stepBtn.Clicked(gtx) {
						pc := sim.GetRegister(datatypes.RegPC)
						instWord := sim.Mem.Work[pc]
						sim.SetRegister(datatypes.RegRI, instWord)
						inst, ifound := sim.InstructionFromWord(instWord)
						if !ifound {
							log.Printf("invalid instruction %x (%d)\n", instWord, instWord)
						} else {
							handler, hfound := sim.Handlers[inst.Repr]
							if !hfound {
								log.Printf("couldn't handle instruction %v\n", inst)
							}
							instPos := datatypes.MachineAddress(pc)
							argsTerm := instPos + 1 + datatypes.MachineAddress(inst.NumArgs)
							// set pc before calling the handler
							// that way branching works
							sim.MapRegister(
								datatypes.RegPC,
								func(pc datatypes.MachineWord) datatypes.MachineWord {
									return pc + datatypes.MachineWord(1+inst.NumArgs)
								},
							)
							args := sim.Mem.Work[instPos:argsTerm]
							log.Printf("Executing %s with %v", inst.Name, args)
							handler(&sim, args)
						}
					}
					return stepBtnView.Layout(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Min}
				}),
			)
		}),
	)
}
