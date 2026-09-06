package grid

import (
	"math"
	"slices"

	i3 "go.i3wm.org/i3/v4"

	"github.com/DanielBrisch/dotfiles/internal/i3x"
)

type Column struct {
	Width   int
	Heights []int
}

func ColumnsFor(count int) int {
	if count < 1 {
		return 1
	}
	return int(math.Ceil(math.Sqrt(float64(count))))
}

func Partition(windows []i3.NodeID, columns int, stackLast bool) [][]i3.NodeID {
	sizes := make([]int, 0, columns)
	remaining := len(windows)
	for column := 0; column < columns; column++ {
		take := int(math.Ceil(float64(remaining) / float64(columns-column)))
		sizes = append(sizes, take)
		remaining -= take
	}
	if stackLast {
		slices.Reverse(sizes)
	}

	result := make([][]i3.NodeID, 0, columns)
	start := 0
	for _, take := range sizes {
		if take == 0 {
			continue
		}
		result = append(result, windows[start:start+take])
		start += take
	}
	return result
}

func CurrentGrid(workspace *i3.Node) [][]i3.NodeID {
	current := make([][]i3.NodeID, 0, len(workspace.Nodes))
	for _, node := range workspace.Nodes {
		if node.Window != 0 {
			current = append(current, []i3.NodeID{node.ID})
			continue
		}
		ids := make([]i3.NodeID, 0, len(node.Nodes))
		for _, leaf := range i3x.Leaves(node) {
			ids = append(ids, leaf.ID)
		}
		current = append(current, ids)
	}
	return current
}

func Proportions(workspace *i3.Node) []Column {
	columns := make([]Column, 0, len(workspace.Nodes))
	for _, node := range workspace.Nodes {
		column := Column{Width: percent(node.Rect.Width, workspace.Rect.Width)}
		if node.Window == 0 {
			for _, leaf := range i3x.Leaves(node) {
				column.Heights = append(column.Heights, percent(leaf.Rect.Height, node.Rect.Height))
			}
		}
		columns = append(columns, column)
	}
	return columns
}

func sameGrid(wanted, current [][]i3.NodeID) bool {
	return slices.EqualFunc(wanted, current, func(a, b []i3.NodeID) bool {
		return slices.Equal(a, b)
	})
}

func percent(part, whole int64) int {
	if whole == 0 {
		whole = 1
	}
	return int(math.Round(float64(part) * 100 / float64(whole)))
}
