package main

import (
	"fmt"
	"strconv"
	"dubcc/datatypes"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	Table struct {
		widget widget.List
		columns []ColumnEnum
		data []TableEntry
	}

	RegisterTableEntry struct {
		name string
		reg *datatypes.Register
	}
	MemoryTableEntry struct {
		address datatypes.MachineAddress
		value datatypes.MachineWord
	}
	TableEntry interface {
		GetColumn(ColumnEnum) string
	}

	ColumnEnum = byte
)

const (
	ColumnName = iota
	ColumnAddress
	ColumnValue
	ColumnBinaryValue
	ColumnMax
)

func (e *MemoryTableEntry) GetColumn(col ColumnEnum) string {
	switch col {
	case ColumnAddress:
		return strconv.FormatUint(uint64(e.address), 10)
	case ColumnValue:
		return strconv.FormatUint(uint64(e.value), 10)
	case ColumnBinaryValue:
		return fmt.Sprintf("%016b", e.value)
	default: return "n/a"
	}
}

func (e *RegisterTableEntry) GetColumn(col ColumnEnum) string {
	switch col {
	case ColumnName:
		return e.name
	case ColumnValue:
		return strconv.FormatUint(uint64(e.reg.Content), 10)
	case ColumnBinaryValue:
		return fmt.Sprintf("%b", e.reg.Content)
	default: return "n/a"
	}
}

var (
	tableMemory = Table {
		widget: widget.List{List: layout.List{Axis: layout.Vertical}},
		columns: []ColumnEnum { ColumnAddress, ColumnValue, ColumnBinaryValue },
	}
	tableRegisters = Table {
		widget: widget.List{List: layout.List{Axis: layout.Vertical}},
		columns: []ColumnEnum { ColumnName, ColumnValue },
	}

	tableColumnNames = map[ColumnEnum]string {
		ColumnName: "Nome",
		ColumnAddress: "Endereço",
		ColumnValue: "Valor",
		ColumnBinaryValue: "Binário",
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

func (tbl *Table) Draw(gtx layout.Context, th *material.Theme, colWeights []float32) layout.Dimensions {
	return material.List(th, &tbl.widget).Layout(gtx, len(tbl.data) + 1, func(gtx layout.Context, i int) layout.Dimensions {
		rowBg := white

		if i > 0 && i%2 != 0 {
			rowBg = zebraColor
		} else if i == 0 {
			rowBg = headerBg
		}

		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: rowBg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		children := make([]layout.FlexChild, len(tbl.columns))

		for j, col := range tbl.columns {
			fontWeight := font.Normal
			if i == 0 {
				fontWeight = font.Bold
			}
			var cellText string
			if i == 0 {
				cellText = tableColumnNames[col]
			} else {
				cellText = tbl.data[i-1].GetColumn(col)
			}
			children[j] = layout.Flexed(colWeights[j], func(gtx layout.Context) layout.Dimensions {
				return drawCell(gtx, th, cellText, fontWeight)
			})
		}

		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	})
}

func InitTables(sim *datatypes.Sim) {
	for name, reg := range sim.Isa.Registers {
		tableRegisters.data = append(
			tableRegisters.data,
			&RegisterTableEntry { name, reg },
		)
	}
	for idx, val := range sim.Mem.Work {
		tableMemory.data = append(
			tableMemory.data,
			&MemoryTableEntry { datatypes.MachineAddress(idx), val },
		)
	}
}
