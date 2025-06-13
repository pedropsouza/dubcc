package main

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
)

func textLayout(gtx layout.Context, th *material.Theme, title string) layout.Dimensions {
	inset := layout.Inset{Left: 8, Right: 8}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
				return FillWithLabel(gtx, th, title, red, 16)
			}),
			layout.Flexed(0.85, func(gtx layout.Context) layout.Dimensions {
				return FillWithText(gtx, th, "Digite aqui...", white)
			}),
		)
	})
}

func TextWithTable(gtx layout.Context, th *material.Theme, title string, bg color.NRGBA, list *widget.List, table [][]string, colWeights []float32) layout.Dimensions {
	inset := layout.Inset{Left: 8, Right: 8}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
				return FillWithLabel(gtx, th, title, red, 16)
			}),
			layout.Flexed(0.85, func(gtx layout.Context) layout.Dimensions {
				ColorBox(gtx, gtx.Constraints.Max, bg)
				return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return drawTable(gtx, th, list, table, colWeights)
				})
			}),
		)
	})
}

func FillWithText(gtx layout.Context, th *material.Theme, text string, bg color.NRGBA) layout.Dimensions {
	ColorBox(gtx, gtx.Constraints.Max, bg)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.Editor(th, &editor, text).Layout(gtx)
		}),
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
