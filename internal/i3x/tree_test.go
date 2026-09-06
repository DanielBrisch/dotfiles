package i3x

import (
	"testing"

	i3 "go.i3wm.org/i3/v4"
)

func TestLeavesIgnoraFlutuantes(t *testing.T) {
	tiled := &i3.Node{ID: 1, Type: i3.Con, Window: 10}
	floating := &i3.Node{
		ID:   2,
		Type: i3.FloatingCon,
		Nodes: []*i3.Node{
			{ID: 3, Type: i3.Con, Window: 30},
		},
	}
	workspace := &i3.Node{
		ID:            4,
		Type:          i3.WorkspaceNode,
		Name:          "1",
		Nodes:         []*i3.Node{tiled},
		FloatingNodes: []*i3.Node{floating},
	}

	leaves := Leaves(workspace)
	if len(leaves) != 1 || leaves[0].ID != tiled.ID {
		t.Fatalf("Leaves = %v, quero apenas a janela lado a lado (%d)", leaves, tiled.ID)
	}
}

func TestLeavesEmLargura(t *testing.T) {
	workspace := &i3.Node{
		Type: i3.WorkspaceNode,
		Nodes: []*i3.Node{
			{Type: i3.Con, Nodes: []*i3.Node{
				{ID: 1, Type: i3.Con, Window: 11},
				{ID: 2, Type: i3.Con, Window: 12},
			}},
			{ID: 3, Type: i3.Con, Window: 13},
		},
	}

	var got []i3.NodeID
	for _, leaf := range Leaves(workspace) {
		got = append(got, leaf.ID)
	}
	want := []i3.NodeID{3, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("Leaves = %v, quero %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Leaves = %v, quero %v", got, want)
		}
	}
}

func TestWorkspaceOfEFocused(t *testing.T) {
	window := &i3.Node{ID: 7, Type: i3.Con, Window: 70, Focused: true}
	workspace := &i3.Node{ID: 5, Type: i3.WorkspaceNode, Name: "3", Nodes: []*i3.Node{window}}
	output := &i3.Node{ID: 4, Type: i3.OutputNode, Nodes: []*i3.Node{workspace}}
	root := &i3.Node{ID: 1, Type: i3.Root, Nodes: []*i3.Node{output}}

	if got := WorkspaceOf(root, window.ID); got != workspace {
		t.Errorf("WorkspaceOf devolveu %v, quero o workspace 3", got)
	}
	if got := FocusedWorkspace(root); got != workspace {
		t.Errorf("FocusedWorkspace devolveu %v, quero o workspace 3", got)
	}
	if got := FindWorkspace(root, "3"); got != workspace {
		t.Errorf("FindWorkspace devolveu %v, quero o workspace 3", got)
	}
}
