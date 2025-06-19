package main

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image/color"
)

var (
	customFont = font.Typeface("FiraCode-Bold")
	red        = color.NRGBA{R: 160, G: 53, B: 47, A: 0xFF}
	yellow     = color.NRGBA{R: 255, G: 212, B: 125, A: 255}
	white      = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	black      = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	headerBg   = color.NRGBA{R: 224, G: 224, B: 224, A: 255}
	cellBorder = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	zebraColor = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
)

func mainLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Flexed(0.20, func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, " D\n U\n B\n c\n c\n", red, 80)
		}),
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(16)}.Layout,
		),
		layout.Flexed(0.60, func(gtx layout.Context) layout.Dimensions {
			return centerLayout(gtx, th)
		}),
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(16)}.Layout,
		),
		layout.Flexed(0.20, func(gtx layout.Context) layout.Dimensions {
			return rightLayout(gtx, th)
		}),
		layout.Rigid(
			layout.Spacer{Width: unit.Dp(16)}.Layout,
		),
	)
}

func centerLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			colWeights := []float32{0.3, 0.3, 0.4}
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return TextWithTable(gtx, th, "MEMÓRIA", white, &tableMemory, colWeights)
			})
		}),
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			return textLayout(gtx, th, "CÓDIGO")
		}),
		layout.Flexed(0.05, func(gtx layout.Context) layout.Dimensions {
			return ColorBox(gtx, gtx.Constraints.Max, red)
		}),
		layout.Flexed(0.1, func(gtx layout.Context) layout.Dimensions {
			return actionButtonsLayout(gtx, th)
		}),
		layout.Flexed(0.05, func(gtx layout.Context) layout.Dimensions {
			return ColorBox(gtx, gtx.Constraints.Max, red)
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
