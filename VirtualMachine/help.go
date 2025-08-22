package main

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type HelpMenu struct {
	listInst widget.List
	listDir  widget.List
	listRegs widget.List
	closeBtn widget.Clickable

	instructions []string
	registers    []string
	directives   []string
}

func NewHelpMenu() *HelpMenu {
	hm := &HelpMenu{
		listInst: widget.List{
			Scrollbar: widget.Scrollbar{},
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		listDir: widget.List{
			Scrollbar: widget.Scrollbar{},
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		listRegs: widget.List{
			Scrollbar: widget.Scrollbar{},
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
	hm.instructions = []string{
		"add",
		"br",
		"brneg",
		"brpos",
		"brzero",
		"copy",
		"divide",
		"load",
		"mult",
		"read",
		"ret",
		"stop",
		"store",
		"sub",
		"write",
		"push",
		"pop",
		"call",
	}
	hm.registers = []string{
		"PC - Program Counter",
		"SP - Stack Pointer",
		"ACC - Accumulator",
		"MOP - Operation Mode",
		"RI - Instruction Register",
		"RE - Address Register",
		"R0 - Multi-Purpose Register",
		"R1 - Multi-Purpose Register",
	}
	hm.directives = []string{
		"TEM Q COMPLETAR AQUI",
	}

	return hm
}

func (hm *HelpMenu) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	header := func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := material.H5(th, "Help List")
				title.Alignment = text.Start
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &hm.closeBtn, "Close")
				btn.Background = yellow
				btn.Color = black
				btn.CornerRadius = unit.Dp(6)
				btn.Inset = layout.UniformInset(unit.Dp(6))
				btn.Font.Typeface = customFont
				if hm.closeBtn.Clicked(gtx) {
					menuBar.showHelpMenu = false
				}
				return btn.Layout(gtx)
			}),
		)
	}

	listColumn := func(list *widget.List, items []string, title string) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(th, title)
					lbl.Alignment = text.Start
					return lbl.Layout(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return material.List(th, list).Layout(gtx, len(items), func(gtx layout.Context, i int) layout.Dimensions {
						return layout.Inset{
							Top:    unit.Dp(2),
							Bottom: unit.Dp(2),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(th, items[i])
							lbl.Alignment = text.Start
							return lbl.Layout(gtx)
						})
					})
				}),
			)
		}
	}
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(header),
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceAround,
			}.Layout(gtx,
				layout.Flexed(1, listColumn(&hm.listInst, hm.instructions, "Instruções")),
				layout.Flexed(1, listColumn(&hm.listRegs, hm.registers, "Registradores")),
				layout.Flexed(1, listColumn(&hm.listDir, hm.directives, "Diretivas")),
			)
		}),
	)
}

/*
func (hm *HelpMenu) topBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	btnStyle := func(l string, c *widget.Clickable) material.ButtonStyle {
		b := material.Button(th, c, l)
		b.Background = yellow
		b.Color = black
		b.CornerRadius = unit.Dp(6)
		b.Inset = layout.UniformInset(unit.Dp(6))
		b.Font.Typeface = customFont
		return b
	}
*/
