package i3x

import i3 "go.i3wm.org/i3/v4"

func Leaves(node *i3.Node) []*i3.Node {
	if node == nil {
		return nil
	}
	var leaves []*i3.Node
	queue := append([]*i3.Node(nil), node.Nodes...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if len(current.Nodes) == 0 {
			if current.Type == i3.Con {
				leaves = append(leaves, current)
			}
			continue
		}
		queue = append(queue, current.Nodes...)
	}
	return leaves
}

func Workspaces(root *i3.Node) []*i3.Node {
	var found []*i3.Node
	var walk func(*i3.Node)
	walk = func(node *i3.Node) {
		if node.Type == i3.WorkspaceNode {
			found = append(found, node)
			return
		}
		for _, child := range node.Nodes {
			walk(child)
		}
	}
	walk(root)
	return found
}

func FindWorkspace(root *i3.Node, name string) *i3.Node {
	for _, workspace := range Workspaces(root) {
		if workspace.Name == name {
			return workspace
		}
	}
	return nil
}

func WorkspaceOf(root *i3.Node, id i3.NodeID) *i3.Node {
	for _, workspace := range Workspaces(root) {
		if workspace.ID == id || contains(workspace, id) {
			return workspace
		}
	}
	return nil
}

func FocusedWorkspace(root *i3.Node) *i3.Node {
	focused := FindFocused(root)
	if focused == nil {
		return nil
	}
	return WorkspaceOf(root, focused.ID)
}

func contains(node *i3.Node, id i3.NodeID) bool {
	for _, child := range append(append([]*i3.Node(nil), node.Nodes...), node.FloatingNodes...) {
		if child.ID == id || contains(child, id) {
			return true
		}
	}
	return false
}

func FindFocused(node *i3.Node) *i3.Node {
	if node.Focused {
		return node
	}
	for _, child := range append(append([]*i3.Node(nil), node.Nodes...), node.FloatingNodes...) {
		if found := FindFocused(child); found != nil {
			return found
		}
	}
	return nil
}
