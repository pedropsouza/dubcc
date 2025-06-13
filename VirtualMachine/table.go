package main

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	tableMemoryList    = widget.List{List: layout.List{Axis: layout.Vertical}}
	tableRegistersList = widget.List{List: layout.List{Axis: layout.Vertical}}

	tableMemoryData = [][]string{
		{"Endereço", "Valor", "Binário"},
		{"000", "42", "101010"},
		{"001", "37", "100101"},
		{"002", "99", "1100011"},
		{"003", "12", "1100"},
		{"004", "88", "1011000"},
		{"000", "42", "101010"},
		{"001", "37", "100101"},
		{"002", "99", "1100011"},
		{"003", "12", "1100"},
		{"004", "88", "1011000"},
		{"000", "42", "101010"},
		{"001", "37", "100101"},
		{"002", "99", "1100011"},
		{"003", "12", "1100"},
		{"004", "88", "1011000"},
	}

	tableRegistersData = [][]string{
		{"registrador", "valor"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
		{"r1", "0"}, {"r1", "0"}, {"r1", "0"},
	}
)

func drawCell(gtx layout.Context, th *material.Theme, text string, weight font.Weight) layout.Dimensions {
	border := widget.Border{
		Color:        cellBorder,
		CornerRadius: unit.Dp(0),
		Width:        unit.Dp(1),
	}

	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{
			Top:    unit.Dp(4),
			Right:  unit.Dp(6),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(6),
		}
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(th, text)
			label.Font.Weight = weight
			return label.Layout(gtx)
		})
	})
}

func drawTable(gtx layout.Context, th *material.Theme, list *widget.List, data [][]string, colWeights []float32) layout.Dimensions {
	return material.List(th, list).Layout(gtx, len(data), func(gtx layout.Context, i int) layout.Dimensions {
		row := data[i]
		rowBg := white

		if i > 0 && i%2 != 0 {
			rowBg = zebraColor
		} else if i == 0 {
			rowBg = headerBg
		}

		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: rowBg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		children := make([]layout.FlexChild, len(row))
		for j, cellText := range row {
			fontWeight := font.Normal
			if i == 0 {
				fontWeight = font.Bold
			}
			children[j] = layout.Flexed(colWeights[j], func(gtx layout.Context) layout.Dimensions {
				return drawCell(gtx, th, cellText, fontWeight)
			})
		}

		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	})
}
