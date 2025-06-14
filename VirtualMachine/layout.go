package main

import (
	"dubcc/datatypes"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image/color"
	"io"
	"log"
	"os"
	"fmt"
	"os/exec"
	"strings"
	"bytes"
)

var (
	customFont = font.Typeface("FiraCode-Bold")
	red        = color.NRGBA{R: 160, G: 53, B: 47, A: 0xFF}
	yellow     = color.NRGBA{R: 255, G: 212, B: 125, A: 255}
	white      = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	headerBg   = color.NRGBA{R: 224, G: 224, B: 224, A: 255}
	cellBorder = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	zebraColor = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
)

func mainLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, " D\n U\n B\n c\n c\n", red, 80)
		}),
		layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
			return centerLayout(gtx, th)
		}),
		layout.Flexed(0.25, func(gtx layout.Context) layout.Dimensions {
			return rightLayout(gtx, th)
		}),
	)
}

func centerLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	assembleBtnView := material.Button(th, &assembleBtn, "Assemble")
	stepBtnView := material.Button(th, &stepBtn, "Step")
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			colWeights := []float32{0.3, 0.3, 0.4}
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return TextWithTable(gtx, th, "MEMÓRIA", white, &tableMemory, colWeights)
			})
		}),
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			return textLayout(gtx, th, "CÓDIGO")
		}),
		layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
			return ColorBox(gtx, gtx.Constraints.Max, red)
		}),
		layout.Flexed(0.1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex {
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
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
								read_file: for mempos := range sim.Mem.Work {
									for idx := range buf {
										readb, err := reader.ReadByte()
										if err != nil {
											if err == io.EOF {
												break read_file
											}
											log.Fatal("error reading stdin: %v", err)
										}
										buf[idx] = readb
									}
									v := datatypes.MachineWord(buf[0] << 8 + buf[1])
									fmt.Fprintf(os.Stderr, "got word %x (%d) out of %v\n", v, v, buf)
									sim.Mem.Work[mempos] = v
								}
							}
						}()
					}
					return assembleBtnView.Layout(gtx)
				}),
				
				// stepBtn
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					if stepBtn.Clicked(gtx) {
						pc := sim.Isa.Registers["PC"]
						ri := sim.Isa.Registers["RI"]
						//re := sim.Isa.Registers["RE"]
						ri.Content = sim.Mem.Work[pc.Content]
						inst, ifound := sim.Isa.InstructionFromWord(ri.Content)
						if !ifound {
							fmt.Fprintf(os.Stderr,
								"invalid instruction %x (%d)\n", ri.Content, ri.Content,
							)
						}
						handler, hfound := sim.Handlers[inst.Repr]
						if !hfound {
							fmt.Fprintf(os.Stderr, "couldn't handle instruction %v\n", inst)
						}
						instPos := datatypes.MachineAddress(pc.Content)
						argsTerm := instPos + 1 + datatypes.MachineAddress(inst.NumArgs)
						// set pc before calling the handler
						// that way branching works
						pc.Content += datatypes.MachineWord(1 + inst.NumArgs)
						args := sim.Mem.Work[instPos:argsTerm]
						log.Printf("Executing %s with %v", inst.Name, args)
						handler(&sim, args)
					}
					return stepBtnView.Layout(gtx)
				}),
			)
		}),
	)
}

func rightLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(0.9, func(gtx layout.Context) layout.Dimensions {
			colWeights := []float32{0.5, 0.5}
			return TextWithTable(gtx, th, "REGISTRADORES", white, &tableRegisters, colWeights)
		}),
		layout.Flexed(0.1, func(gtx layout.Context) layout.Dimensions {
			return ColorBox(gtx, gtx.Constraints.Max, red)
		}),
	)
}
