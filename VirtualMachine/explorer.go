package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type FileExplorer struct {
	current string
	entries []fs.DirEntry

	list        widget.List
	entryClicks []widget.Clickable
	closeBtn    widget.Clickable
	upBtn       widget.Clickable
	backBtn     widget.Clickable
	refreshBtn  widget.Clickable

	history   []string
	bcClicks  []widget.Clickable // breadcrumbs
	extFilter map[string]struct{}
	OnSelect  func(path string)
}

func NewFileExplorer() *FileExplorer {
	fe := &FileExplorer{
		list: widget.List{List: layout.List{Axis: layout.Vertical}},
	}
	if cwd, err := os.Getwd(); err == nil {
		fe.setDir(cwd, false)
	}
	return fe
}

func (fe *FileExplorer) SetStartDir(path string) {
	fe.setDir(path, false)
}

func (fe *FileExplorer) SetFilter(exts ...string) {
	fe.extFilter = make(map[string]struct{}, len(exts))
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e != "" && e[0] != '.' {
			e = "." + e
		}
		fe.extFilter[e] = struct{}{}
	}
}

func (fe *FileExplorer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return fe.topBar(gtx, th)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return fe.listArea(gtx, th)
		}),
	)
}

func (fe *FileExplorer) topBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	btnStyle := func(l string, c *widget.Clickable) material.ButtonStyle {
		b := material.Button(th, c, l)
		b.Background = yellow
		b.Color = black
		b.CornerRadius = unit.Dp(6)
		b.Inset = layout.UniformInset(unit.Dp(6))
		b.Font.Typeface = customFont
		return b
	}

	if fe.closeBtn.Clicked(gtx) {
		showExplorer = false
	}
	if fe.backBtn.Clicked(gtx) && len(fe.history) > 1 {
		fe.history = fe.history[:len(fe.history)-1]
		prev := fe.history[len(fe.history)-1]
		fe.setDir(prev, true)
	}
	if fe.upBtn.Clicked(gtx) {
		parent := filepath.Dir(fe.current)
		fe.setDir(parent, false)
	}
	if fe.refreshBtn.Clicked(gtx) {
		fe.setDir(fe.current, true)
	}

	segments := splitPathSegments(fe.current)
	if len(fe.bcClicks) < len(segments) {
		for i := len(fe.bcClicks); i < len(segments); i++ {
			fe.bcClicks = append(fe.bcClicks, widget.Clickable{})
		}
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return btnStyle("Close", &fe.closeBtn).Layout(gtx) }),
		layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return btnStyle("Back", &fe.backBtn).Layout(gtx) }),
		layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return btnStyle("Up", &fe.upBtn).Layout(gtx) }),
		layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return btnStyle("Refresh", &fe.refreshBtn).Layout(gtx) }),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					l := material.Label(th, th.TextSize, "Path:")
					l.Color = black
					l.Font.Typeface = customFont
					return l.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						func() []layout.FlexChild {
							var children []layout.FlexChild
							acc := ""
							for i, seg := range segments {
								if i == 0 {
									acc = seg
								} else {
									acc = filepath.Join(acc, seg)
								}
								idx := i
								children = append(children,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										b := material.Button(th, &fe.bcClicks[idx], seg)
										b.Background = white
										b.Color = black
										b.CornerRadius = unit.Dp(6)
										b.Inset = layout.UniformInset(unit.Dp(4))
										b.Font.Typeface = customFont
										if fe.bcClicks[idx].Clicked(gtx) {
											fe.setDir(joinSegments(segments[:idx+1]), false)
										}
										return b.Layout(gtx)
									}),
								)
								if i < len(segments)-1 {
									children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
									children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										l := material.Label(th, th.TextSize, "/")
										l.Color = black
										return l.Layout(gtx)
									}))
									children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
								}
							}
							return children
						}()...,
					)
				}),
			)
		}),
	)
}

func (fe *FileExplorer) listArea(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(fe.entryClicks) != len(fe.entries) {
		fe.entryClicks = make([]widget.Clickable, len(fe.entries))
	}

	row := func(i int) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			de := fe.entries[i]
			name := de.Name()
			isDir := de.IsDir()

			btn := material.Button(th, &fe.entryClicks[i], name)
			btn.Inset = layout.UniformInset(unit.Dp(8))
			btn.CornerRadius = unit.Dp(8)
			btn.Font.Typeface = customFont
			if isDir {
				btn.Background = yellow
				btn.Color = black
			} else {
				btn.Background = white
				btn.Color = black
			}

			// info adicional
			var sub string
			if info, err := de.Info(); err == nil {
				if isDir {
					sub = "DIR • " + info.ModTime().Format("2006-01-02 15:04")
				} else {
					sub = byteSize(info.Size()) + " • " + info.ModTime().Format("2006-01-02 15:04")
				}
			}
			if sub != "" {
				btn.CornerRadius = unit.Dp(8)
			}

			if fe.entryClicks[i].Clicked(gtx) {
				full := filepath.Join(fe.current, name)
				if isDir {
					fe.setDir(full, false)
				} else if fe.OnSelect != nil {
					fe.OnSelect(full)
				}
			}

			return btn.Layout(gtx)
		}
	}

	return material.List(th, &fe.list).Layout(gtx, len(fe.entries), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(2)).Layout(gtx, row(i))
	})
}

func (fe *FileExplorer) setDir(path string, replaceTop bool) {
	abs := path
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
	}
	ents, _ := os.ReadDir(abs)
	// aplica filtro + ordena: dirs primeiro, depois arquivos (alfabético)
	filtered := ents[:0]
	for _, e := range ents {
		if fe.extFilter != nil && len(fe.extFilter) > 0 && !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if _, ok := fe.extFilter[ext]; !ok {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	sort.Slice(filtered, func(i, j int) bool {
		di, dj := filtered[i].IsDir(), filtered[j].IsDir()
		if di != dj {
			return di // diretórios vêm antes
		}
		return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
	})

	fe.entries = filtered
	fe.current = abs

	// histórico
	if len(fe.history) == 0 {
		fe.history = []string{abs}
	} else if replaceTop {
		fe.history[len(fe.history)-1] = abs
	} else if fe.history[len(fe.history)-1] != abs {
		fe.history = append(fe.history, abs)
	}
}

// utilitários

func splitPathSegments(p string) []string {
	if p == "" || p == "/" {
		return []string{"/"}
	}
	vol := filepath.VolumeName(p) // no Windows, "C:"
	p = strings.TrimPrefix(p, vol)
	parts := strings.Split(strings.Trim(p, string(filepath.Separator)), string(filepath.Separator))
	if vol != "" {
		return append([]string{vol + string(filepath.Separator)}, parts...)
	}
	if filepath.IsAbs(p) {
		return append([]string{string(filepath.Separator)}, parts...)
	}
	return parts
}

func joinSegments(segs []string) string {
	if len(segs) == 0 {
		return string(filepath.Separator)
	}
	// trata raiz
	if segs[0] == string(filepath.Separator) || strings.HasSuffix(segs[0], string(filepath.Separator)) {
		return filepath.Join(segs...)
	}
	return filepath.Join(segs...)
}

func byteSize(n int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
	)
	f := float64(n)
	switch {
	case f >= GB:
		return sprintf1("%.2f GB", f/GB)
	case f >= MB:
		return sprintf1("%.2f MB", f/MB)
	case f >= KB:
		return sprintf1("%.2f KB", f/KB)
	default:
		return sprintf1("%d B", n)
	}
}

func sprintf1(format string, a any) string {
	return strings.TrimSuffix(strings.TrimSuffix(time.Now().Format(""), ""), "")
}
