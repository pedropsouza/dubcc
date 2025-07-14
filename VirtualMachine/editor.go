package main

import (
	"fmt"
	"image"
	"image/color"
	//"log"
	_ "net/http/pprof" // This line registers the pprof handlers
	//"os"
	"regexp"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/oligo/gvcode"
	gvcolor "github.com/oligo/gvcode/color"
	"github.com/oligo/gvcode/textstyle/syntax"
	//wg "github.com/oligo/gvcode/widget"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type EditorApp struct {
	state   *gvcode.Editor
	xScroll widget.Scrollbar
	yScroll widget.Scrollbar
}

const (
	syntaxPattern = "package|import|type|func|struct|for|var|switch|case|if|else"
)

func (ed *EditorApp) Layout(gtx C, th *material.Theme) D {
	for {
		evt, ok := ed.state.Update(gtx)
		if !ok {
			break
		}

		switch evt.(type) {
		case gvcode.ChangeEvent:
			tokens := HightlightTextByPattern(ed.state.Text(), syntaxPattern)
			ed.state.SetSyntaxTokens(tokens...)
		}
	}

	xScrollDist := ed.xScroll.ScrollDistance()
	yScrollDist := ed.yScroll.ScrollDistance()
	if xScrollDist != 0.0 || yScrollDist != 0.0 {
		ed.state.Scroll(gtx, xScrollDist, yScrollDist)
	}

	scrollIndicatorColor := gvcolor.MakeColor(th.Fg).MulAlpha(0x30)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{
				Top:   unit.Dp(2),
				Left:  unit.Dp(6),
				Right: unit.Dp(6),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Horizontal,
				}.Layout(gtx,
					layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
						ed.state.WithOptions(
							gvcode.WithFont(font.Font{Typeface: "monospace", Weight: font.SemiBold}),
							gvcode.WithTextSize(unit.Sp(12)),
							gvcode.WithLineHeight(0, 1.5),
						)

						dims := ed.state.Layout(gtx, th.Shaper)

						macro := op.Record(gtx.Ops)
						scrollbarDims := func(gtx C) D {
							return layout.Inset{
								Left: gtx.Metric.PxToDp(ed.state.GutterWidth()),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								minX, maxX, _, _ := ed.state.ScrollRatio()
								bar := makeScrollbar(th, &ed.xScroll, scrollIndicatorColor.NRGBA())
								return bar.Layout(gtx, layout.Horizontal, minX, maxX)
							})
						}(gtx)

						scrollbarOp := macro.Stop()
						defer op.Offset(image.Point{Y: dims.Size.Y - scrollbarDims.Size.Y}).Push(gtx.Ops).Pop()
						scrollbarOp.Add(gtx.Ops)

						return dims
					}),

					layout.Rigid(func(gtx C) D {
						_, _, minY, maxY := ed.state.ScrollRatio()
						bar := makeScrollbar(th, &ed.yScroll, scrollIndicatorColor.NRGBA())
						return bar.Layout(gtx, layout.Vertical, minY, maxY)
					}),
				)

			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Right:  unit.Dp(8),
				Top:    unit.Dp(2),
				Bottom: unit.Dp(2),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				line, col := ed.state.CaretPos()
				lb := material.Label(th, th.TextSize*0.7, fmt.Sprintf("Line:%d, Col:%d", line+1, col+1))
				lb.Alignment = text.End
				lb.Color = ed.state.ColorPalette().Foreground.NRGBA()
				return lb.Layout(gtx)
			})
		}),
	)

}

func makeScrollbar(th *material.Theme, scroll *widget.Scrollbar, color color.NRGBA) material.ScrollbarStyle {
	bar := material.Scrollbar(th, scroll)
	bar.Indicator.Color = color
	bar.Indicator.CornerRadius = unit.Dp(0)
	bar.Indicator.MinorWidth = unit.Dp(12)
	bar.Track.MajorPadding = unit.Dp(0)
	bar.Track.MinorPadding = unit.Dp(1)
	return bar
}

func HightlightTextByPattern(text string, pattern string) []syntax.Token {
	var tokens []syntax.Token

	re := regexp.MustCompile(pattern)
	matches := re.FindAllIndex([]byte(text), -1)
	for _, match := range matches {
		tokens = append(tokens, syntax.Token{
			Start: match[0],
			End:   match[1],
			Scope: "keyword",
		})
	}

	return tokens
}
