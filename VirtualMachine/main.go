package main

import (
	"bytes"
	"dubcc"
	"dubcc/assembler"
	"dubcc/linker"
	_ "embed"
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
	"time"
)

type LinkerMode = linker.LinkerMode
type MachineAddress = dubcc.MachineAddress
type SourceFile = assembler.SourceFile

const (
	Relocator = linker.Relocator
	Absolute = linker.Absolute
)

var files	[]SourceFile
var linkerMode LinkerMode
var loadAddress MachineAddress
var executableProvided bool = false
var window *app.Window
var editor EditorApp
var th *material.Theme
var assembleBtn, stepBtn, resetBtn, runBtn, stopBtn widget.Clickable
var menuBar MenuBar
var hexView = false
var terminal *Terminal
var loopTimer *time.Ticker
var loopChan chan bool = make(chan bool)

var showExplorer bool
var fe = NewFileExplorer()

var showSaveDialog bool
var saveExplorer *FileExplorer
var filenameEditor widget.Editor
var saveConfirmBtn widget.Clickable
var saveCancelBtn widget.Clickable
var showHelpMenu bool
var helpMenu = NewHelpMenu()
var currentFilename string

var memCap dubcc.MachineAddress
var sim dubcc.Sim
var assemblerSingleton assembler.Info

//go:embed appicon.png
var logoData []byte

type MenuBar struct {
	fileBtn, hexBtn, editBtn, helpBtn                widget.Clickable
	openBtn, saveBtn, saveAsBtn, exitBtn     widget.Clickable
	showFileMenu, showEditMenu, showHelpMenu bool
	menuWidth                                int
	backdrop                                 widget.Clickable
}

func main() {
	memCap = dubcc.MachineAddress(1 << 10)
	sim = dubcc.MakeSim(memCap)
	assemblerSingleton = assembler.MakeAssembler()
	InitTables(&sim)
	editor = EditorApp{}
	th = material.NewTheme()
	editor.state = wg.NewEditor(th)
	gvcode.SetDebug(false)

	// commandline arguments parsing
	if len(os.Args) > 1 {
		sourceAlreadyRead := false
		linkerMode = Relocator
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			switch arg {
			case "-e", "--executable":
				if len(os.Args) != i+1 {
					obj, err := os.ReadFile(os.Args[i+1])
					if err != nil {
						log.Printf("%v\n", err)
						continue
					}
					r := bytes.NewReader(obj)
					executable, err := assembler.Read(r)
					if err != nil {
						panic(err.Error())
					}
					file := SourceFile{
						Name: string(obj),
						Data: "",
						Object: executable,
					}
					files = append(files, file)
					executableProvided = true
					continue
				} else {
					log.Fatal("error: --executable needs <executabe path>")
				}
			case "-l", "--lst":
				log.Fatal("not implemented xd")
			case "-a", "--absolute":
				if len(os.Args) != i+1 {
					log.Fatal("usage: --absolute <load address>")
				} else {
					linkerMode = Absolute
					var err error
					loadAddress, err = assembler.ParseNum(os.Args[i+1])
					if err != nil {
						log.Fatal("error: " + err.Error())
					}
					continue
				} 
			case "-s", "--save-temps":
				sim.SaveTemps = true
			default:
				code, err := os.ReadFile(arg)
				if err != nil {
					log.Printf("%v\n", err)
					continue
				}
				file := SourceFile{
					Name: string(arg),
					Data: string(code),
					Object: nil,
				}
				files = append(files, file)
				if !sourceAlreadyRead {
					editor.state.SetText(files[0].Data)
					sourceAlreadyRead = true
				}
			}
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
	terminal = NewTerminal(th)
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

			select {
			case <- loopChan:
				StepSimulation()
			default:
			}


			paint.ColorOp{Color: red}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)

			mainLayout(gtx, th)

			e.Frame(gtx.Ops)
		}
	}
}

var logoWidget widget.Image

func init() {
	img, err := png.Decode(bytes.NewReader(logoData))
	if err != nil {
		log.Fatalf("Falha ao decodificar a imagem: %v", err)
	}
	fe.SetStartDir(".")
	fe.OnSelect = func(path string) {
		data, err := os.ReadFile(path)
		if err == nil {
			editor.state.SetText(string(data))
		}
		showExplorer = false
	}
	saveExplorer = NewFileExplorer()
	saveExplorer.SetStartDir(".")
	saveExplorer.OnSelect = nil
	logoWidget = widget.Image{
		Src: paint.NewImageOp(img),
		Fit: widget.Contain,
	}
}
