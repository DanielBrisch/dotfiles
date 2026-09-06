package grid

import (
	"slices"
	"testing"

	i3 "go.i3wm.org/i3/v4"
)

func ids(count int) []i3.NodeID {
	windows := make([]i3.NodeID, count)
	for index := range windows {
		windows[index] = i3.NodeID(index + 1)
	}
	return windows
}

func shape(columns [][]i3.NodeID) []int {
	sizes := make([]int, len(columns))
	for index, column := range columns {
		sizes[index] = len(column)
	}
	return sizes
}

func TestColumnsFor(t *testing.T) {
	want := map[int]int{0: 1, 1: 1, 2: 2, 3: 2, 4: 2, 5: 3, 9: 3, 10: 4}
	for count, expected := range want {
		if got := ColumnsFor(count); got != expected {
			t.Errorf("ColumnsFor(%d) = %d, quero %d", count, got, expected)
		}
	}
}

func TestPartitionShape(t *testing.T) {
	cases := []struct {
		count  int
		normal []int
		flip   []int
	}{
		{1, []int{1}, []int{1}},
		{2, []int{1, 1}, []int{1, 1}},
		{3, []int{2, 1}, []int{1, 2}},
		{4, []int{2, 2}, []int{2, 2}},
		{5, []int{2, 2, 1}, []int{1, 2, 2}},
		{7, []int{3, 2, 2}, []int{2, 2, 3}},
	}
	for _, test := range cases {
		windows := ids(test.count)
		columns := ColumnsFor(test.count)
		if got := shape(Partition(windows, columns, false)); !slices.Equal(got, test.normal) {
			t.Errorf("Partition(%d, false) = %v, quero %v", test.count, got, test.normal)
		}
		if got := shape(Partition(windows, columns, true)); !slices.Equal(got, test.flip) {
			t.Errorf("Partition(%d, true) = %v, quero %v", test.count, got, test.flip)
		}
	}
}

func TestPartitionPreservaJanelas(t *testing.T) {
	for count := 1; count <= 12; count++ {
		windows := ids(count)
		for _, stackLast := range []bool{false, true} {
			var flat []i3.NodeID
			for _, column := range Partition(windows, ColumnsFor(count), stackLast) {
				flat = append(flat, column...)
			}
			if !slices.Equal(flat, windows) {
				t.Errorf("Partition(%d, %v) perdeu ou reordenou janelas: %v", count, stackLast, flat)
			}
		}
	}
}

func TestSameGrid(t *testing.T) {
	wanted := [][]i3.NodeID{{1, 2}, {3}}
	if !sameGrid(wanted, [][]i3.NodeID{{1, 2}, {3}}) {
		t.Error("grades iguais foram consideradas diferentes")
	}
	if sameGrid(wanted, [][]i3.NodeID{{1}, {2, 3}}) {
		t.Error("grades diferentes foram consideradas iguais")
	}
}

func TestProportions(t *testing.T) {
	workspace := &i3.Node{
		Rect: i3.Rect{Width: 1000, Height: 500},
		Nodes: []*i3.Node{
			{
				Rect: i3.Rect{Width: 300, Height: 500},
				Nodes: []*i3.Node{
					{Type: i3.Con, Rect: i3.Rect{Height: 200}, Window: 1},
					{Type: i3.Con, Rect: i3.Rect{Height: 300}, Window: 2},
				},
			},
			{Rect: i3.Rect{Width: 700, Height: 500}, Window: 3},
		},
	}

	got := Proportions(workspace)
	want := []Column{{Width: 30, Heights: []int{40, 60}}, {Width: 70}}
	if len(got) != len(want) {
		t.Fatalf("Proportions devolveu %d colunas, quero %d", len(got), len(want))
	}
	for index, column := range got {
		if column.Width != want[index].Width || !slices.Equal(column.Heights, want[index].Heights) {
			t.Errorf("coluna %d = %+v, quero %+v", index, column, want[index])
		}
	}
}
