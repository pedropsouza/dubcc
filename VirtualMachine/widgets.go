package main

import (
	"dubcc"
	"dubcc/assembler"
	"dubcc/linker"
	"dubcc/macroprocessor"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/k0kubun/pp/v3"
)

type Linker = linker.Linker
type ObjectFile = assembler.ObjectFile

func (mb *MenuBar) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	bar := layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mb.fileBtn.Clicked(gtx) {
				mb.showFileMenu = !mb.showFileMenu
			}
			btn := FlatButton(th, &mb.fileBtn, "File")
			btn.Background = yellow
			btn.Color = black
			btn.Font.Typeface = customFont
			return btn.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mb.hexBtn.Clicked(gtx) {
				hexView = !hexView
			}
			btnText := "Hex" // Texto padrão
			if hexView {
				btnText = "List"
			}

			btn := FlatButton(th, &mb.hexBtn, btnText)
			btn.Background = yellow
			btn.Color = black
			btn.Font.Typeface = customFont
			return btn.Layout(gtx)
		}),

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
				if currentFilename != "" {
					err := os.WriteFile(currentFilename, []byte(editor.state.Text()), 0644)
					if err != nil {
						log.Printf("Error: couldn't save file: %v", err)
					} else {
						log.Printf("File saved succesfully: %s", currentFilename)
					}
				} else {
					renderSaveAsDialog(gtx, th)
				}
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
		currentFilename = fullPath
	}
}

func textLayout(gtx layout.Context, th *material.Theme, title string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, title, red, 16)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return FillWithText(gtx, th, "Digite aqui...", white)
		}),
	)
}

func TextWithTable(gtx layout.Context, th *material.Theme, title string, bg color.NRGBA, table *Table, colWeights []float32) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, title, red, 16)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return table.Draw(gtx, th, colWeights)
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
					startAddressMachineWord := dubcc.MachineWord(assemblerSingleton.StartAddress)
					sim.SetRegister(dubcc.RegPC, startAddressMachineWord)
					terminal.Clear()
					WipeMemory()
				}
				return resetBtnView.Layout(gtx)
			}
		}),
	)
}

func CompileCode() {
	terminal.Clear()
	sim.Registers = dubcc.StartupRegisters(&sim.Isa, dubcc.MachineAddress(len(sim.Mem.Work)))
	if len(files) < 1 {
		files = append(files, SourceFile{Name: "editor", Data: ""})
	}
	files[0].Data = editor.state.Text()
	print(files)
	// we definetely should make this function smaller
	sim.Registers = dubcc.StartupRegisters(&sim.Isa, dubcc.MachineAddress(len(sim.Mem.Work)))
	var assemblers = make([]assembler.Info, len(files))
	var linkerSingleton *Linker
	var objects []*assembler.ObjectFile

	if executableProvided { goto populateMemory }

	switch linkerMode {
	case Relocator:
		linkerSingleton = linker.MakeRelocatorLinker()
	case Absolute:
		linkerSingleton = linker.MakeAbsoluteLinker(0)
	}

	for i := range files {
		macroProcessor := macroprocessor.MakeMacroProcessor(0)
		expanded := []string{}
    for _, line := range strings.Split(files[i].Data, "\n") {
			lines, err := macroProcessor.ProcessLine(line)
			if err != nil {
				log.Print(err)
			}
			expanded = append(expanded, lines...)
    }

    asm := assembler.MakeAssembler()
		{
			// I believe this should be generated after the linking etc
			fname := files[i].Name
			fname_parts := strings.Split(fname, "/")
			fname = fname_parts[len(fname_parts)-1]
			fname = fmt.Sprintf("MASMAPRG-%s.ASM", fname)
			masmaprg, err := os.Create(fname)
			if err != nil {
				log.Printf("couldn't create macro expansion file %s! Ignoring.",
				fname)
			} else {
				defer masmaprg.Close()
			}
			for _, line := range expanded {
				masmaprg.WriteString(line + "\n")
				asm.FirstPassString(line)
			}
		}

		println(macroProcessor.MacroReport())

    _, err := asm.SecondPass()
		if err != nil {
			err = fmt.Errorf("error: could not compile: %v", err)
			terminal.Write(err.Error())
			return
		}
    assemblers = append(assemblers, asm)

		obj, err := asm.GenerateObjectFile()
		if err != nil {
			err = fmt.Errorf("error: could not compile: %v", err)
			terminal.Write(err.Error())
			return
		}
		pp.Print(obj)

		files[i].Object = obj
		log.Printf("code compiled successfully: %d symbols, %d relocations",
			len(obj.Symbols), len(obj.Relocations))
		
		if sim.SaveTemps {
			path := files[i].Name
			base := filepath.Base(path)
			if dot := strings.LastIndex(base, "."); dot != -1 {
				base = base[:dot]
			}
			objFilename := base + ".o"
			if err := assembler.SaveCompleteObjectFile(obj, objFilename); err != nil {
				log.Printf("warning: could not save %s: %v", objFilename, err)
			}
		}
	}

	for i := range files {
		objects = append(objects, files[i].Object)
	}

populateMemory:
	var executable *ObjectFile

	if  executableProvided {
		executable = files[0].Object
	} else {
		var err error
		executable, err = linkerSingleton.GenerateExecutable(objects)
		if err != nil {
			log.Printf("error: could not generate an executable\n%s\n", err.Error())
		}
	}

	mem := []dubcc.MachineWord{}

	for _, section := range executable.Sections {
		addr := section.Header.Address
		tail := len(mem)
		if tail < int(addr) {
			mem = slices.Grow(mem, int(addr) - tail)
		}
		if int(addr) < tail {
			panic("section overlap")
		}
		mem = append(mem, section.Data...)
	}

	print(executable.PrettyPrint())

	if len(mem) > len(sim.Mem.Work) {
		panic("program's too big")
	}

	startAddressMachineWord := dubcc.MachineWord(assemblers[0].StartAddress)
	sim.SetRegister(dubcc.RegPC, startAddressMachineWord) // Altera o valor do PC pro valor indicado na diretiva "start"
	startAddressInt := int(assemblers[0].StartAddress)

	// NOTE: loader starts here
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
		oldPc := sim.GetRegister(dubcc.RegPC)
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
		switch sim.State {
		case dubcc.SimStateRun: sim.State = dubcc.SimStatePause
		case dubcc.SimStateIOBlocked:
		  sim.SetRegister(dubcc.RegPC, oldPc) // actually block
		}

		if len(sim.OutWords) > 0 {
			terminal.Write(string(sim.RxOutWord()))
		}
	}
}

func WipeMemory() {
	memCap := len(sim.Mem.Work)
	for i := range memCap {
		sim.Mem.Work[i] = 0
	}
}

func FlatButton(th *material.Theme, btn *widget.Clickable, label string) material.ButtonStyle {
	b := material.Button(th, btn, label)
	b.Background = yellow
	b.Color = black
	b.Inset = layout.UniformInset(unit.Dp(4)) // deixa mais fino (default é ~8)
	b.CornerRadius = unit.Dp(0)               // deixa quadrado
	b.Font.Typeface = customFont
	return b
}

const (
	addressWidthDP = unit.Dp(70)
	hexCellWidthDP = unit.Dp(55)
)

type HexViewer struct {
	widget       widget.List
	bytesPerRow  int
	startAddress dubcc.MachineAddress
	data         []dubcc.MachineWord
	addressWidth int
}

func NewHexViewer() *HexViewer {
	return &HexViewer{
		widget:       widget.List{List: layout.List{Axis: layout.Vertical}},
		bytesPerRow:  32, // Máximo de bytes por linha
		startAddress: 0,
		addressWidth: 4,
	}
}

func (hv *HexViewer) SetData(data []dubcc.MachineWord, startAddr dubcc.MachineAddress) {
	hv.data = data
	hv.startAddress = startAddr
}

func (hv *HexViewer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(hv.data) == 0 {
		return layout.Dimensions{}
	}

	availableWidth := gtx.Constraints.Max.X - gtx.Dp(addressWidthDP)
	bytesPerRow := availableWidth / gtx.Dp(hexCellWidthDP)

	if bytesPerRow < 4 {
		bytesPerRow = 4
	}
	if bytesPerRow > hv.bytesPerRow {
		bytesPerRow = hv.bytesPerRow
	}

	numRows := (len(hv.data) + bytesPerRow - 1) / bytesPerRow

	return material.List(th, &hv.widget).Layout(gtx, numRows, func(gtx layout.Context, rowIndex int) layout.Dimensions {
		return hv.drawRow(gtx, th, rowIndex, bytesPerRow)
	})
}

func (hv *HexViewer) drawRow(gtx layout.Context, th *material.Theme, rowIndex int, bytesPerRow int) layout.Dimensions {
	startIdx := rowIndex * bytesPerRow
	endIdx := startIdx + bytesPerRow
	if endIdx > len(hv.data) {
		endIdx = len(hv.data)
	}

	rowBg := zebraColor
	if rowIndex%2 == 1 {
		rowBg = white
	}

	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: rowBg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			addr := hv.startAddress + dubcc.MachineAddress(startIdx)
			addrText := fmt.Sprintf("%0*X:", hv.addressWidth, addr)
			return hv.drawCell(gtx, th, addrText, font.Bold, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {

			return hv.drawHexData(gtx, th, startIdx, endIdx)
		}),
	)
}

func (hv *HexViewer) drawHexData(gtx layout.Context, th *material.Theme, startIdx, endIdx int) layout.Dimensions {
	var children []layout.FlexChild

	for i := startIdx; i < endIdx; i++ {
		word := hv.data[i]
		hexText := fmt.Sprintf("%04X", word)

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return hv.drawCell(gtx, th, hexText, font.Normal, black)
		}))
	}
	flex := layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceAround,
	}
	return flex.Layout(gtx, children...)
}

func (hv *HexViewer) drawCell(gtx layout.Context, th *material.Theme, text string, weight font.Weight, textColor color.NRGBA) layout.Dimensions {
	inset := layout.Inset{
		Top:    unit.Dp(6),
		Right:  unit.Dp(6),
		Bottom: unit.Dp(6),
		Left:   unit.Dp(6),
	}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Body1(th, text)
		label.Font.Weight = weight
		label.Font.Typeface = customFont
		label.TextSize = unit.Sp(18) // Retornado ao tamanho original
		label.Color = textColor
		label.MaxLines = 1
		return label.Layout(gtx)
	})
}

func HexViewerWithTitle(gtx layout.Context, th *material.Theme, title string, hexViewer *HexViewer) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, title, red, 16)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {

				return hexViewer.Layout(gtx, th)
			})
		}),
	)
}

var hexViewer = NewHexViewer()

func UpdateHexViewer() {
	memorySlice := sim.Mem.Work[:min(int(memCap), len(sim.Mem.Work))]
	hexViewer.SetData(memorySlice, 0)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
