package main

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
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
	return layout.Stack{}.Layout(gtx,
		// Camada base: sua UI normal (barra de menu + conteúdo)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return menuBar.Layout(gtx, th)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(
						gtx,
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
						layout.Flexed(0.70, func(gtx layout.Context) layout.Dimensions { return centerLayout(gtx, th) }),
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
						layout.Flexed(0.30, func(gtx layout.Context) layout.Dimensions { return rightLayout(gtx, th) }),
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					)
				}),
			)
		}),

		// Camada 2: backdrop clicável (só quando o menu está aberto)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !menuBar.showFileMenu {
				return layout.Dimensions{}
			}
			// se clicar fora, fecha
			if menuBar.backdrop.Clicked(gtx) {
				menuBar.showFileMenu = false
			}
			// clickable invisível cobrindo a tela toda
			return menuBar.backdrop.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
		}),

		// Camada 3: o dropdown em si (desenhado acima do backdrop)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !menuBar.showFileMenu {
				return layout.Dimensions{}
			}
			// posicione o dropdown: um pouco à direita e logo abaixo da barra
			// (ajuste estes números conforme sua barra)
			x := gtx.Dp(unit.Dp(8))
			menuH := gtx.Dp(unit.Dp(36))
			op := op.Offset(image.Pt(x, menuH))
			defer op.Push(gtx.Ops).Pop()

			return menuBar.renderFileMenu(gtx, th)
		}),
	)
}

func centerLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			colWeights := []float32{0.2, 0.15, 0.35, 0.3}
			return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return TextWithTable(gtx, th, "MEMÓRIA", red, &tableMemory, colWeights)
			})
		}),
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			return textLayout(gtx, th, "CÓDIGO")
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
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				colWeights := []float32{0.33, 0.33, 0.33}
				return TextWithTable(gtx, th, "REGISTRADORES", red, &tableRegisters, colWeights)
			}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(32)}.Layout,
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return actionButtonsLayout(gtx, th)
		}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(32)}.Layout,
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return logoWidget.Layout(gtx) }),
	)
}
