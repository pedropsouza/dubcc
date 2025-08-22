package main

import (
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return menuBar.Layout(gtx, th)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(
						gtx,
						layout.Flexed(0.70, func(gtx layout.Context) layout.Dimensions {
							// Adiciona padding nas laterais do layout central
							inset := layout.Inset{Left: unit.Dp(16), Right: unit.Dp(8)}
							return inset.Layout(gtx, func(gtx C) D {
								return centerLayout(gtx, th)
							})
						}),
						layout.Flexed(0.30, func(gtx layout.Context) layout.Dimensions {
							// Adiciona padding nas laterais do layout direito
							inset := layout.Inset{Left: unit.Dp(8), Right: unit.Dp(16)}
							return inset.Layout(gtx, func(gtx C) D {
								return rightLayout(gtx, th)
							})
						}),
					)
				}),
			)
		}),

		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !menuBar.showFileMenu {
				return layout.Dimensions{}
			}
			if menuBar.backdrop.Clicked(gtx) {
				menuBar.showFileMenu = false
			}
			return menuBar.backdrop.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
		}),

		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !menuBar.showFileMenu {
				return layout.Dimensions{}
			}
			x := gtx.Dp(unit.Dp(8))
			menuH := gtx.Dp(unit.Dp(36))
			op := op.Offset(image.Pt(x, menuH))
			defer op.Push(gtx.Ops).Pop()

			return menuBar.renderFileMenu(gtx, th)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !showExplorer {
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
				return fe.Layout(gtx2, th)
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if !menuBar.showHelpMenu {
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
				return helpMenu.Layout(gtx, th)
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return renderSaveAsDialog(gtx, th)
		}),
	)
}

func centerLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			if hexView {
				UpdateHexViewer() // Update with current memory state
				return HexViewerWithTitle(gtx, th, "MEMORY HEX VIEW", hexViewer)
			} else {
				colWeights := []float32{0.2, 0.15, 0.35, 0.3}
				return layout.Inset{
					Left: unit.Dp(8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return TextWithTable(gtx, th, "MEMORY", red, &tableMemory, colWeights)
				})
			}
		}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(16)}.Layout,
		),
		layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
			return textLayout(gtx, th, "CODE")
		}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(16)}.Layout,
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return actionButtonsLayout(gtx, th)
		}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(16)}.Layout,
		),
	)
}

func rightLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				colWeights := []float32{0.33, 0.33, 0.33}
				return TextWithTable(gtx, th, "REGISTERS", red, &tableRegisters, colWeights)
			}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return LayoutGeral(gtx, terminal)
			})
		}),
		layout.Rigid(
			layout.Spacer{Height: unit.Dp(16)}.Layout,
		),
		//layout.Rigid(func(gtx layout.Context) layout.Dimensions { return logoWidget.Layout(gtx) }),
	)
}
