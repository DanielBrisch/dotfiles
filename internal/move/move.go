package move

import (
	"fmt"
	"slices"

	i3 "go.i3wm.org/i3/v4"

	"github.com/DanielBrisch/dotfiles/internal/i3x"
)

var Directions = []string{"left", "right", "up", "down"}

func Run(direction string) error {
	if !slices.Contains(Directions, direction) {
		return fmt.Errorf("direcao invalida: %s", direction)
	}

	tree, err := i3.GetTree()
	if err != nil {
		return err
	}
	focused := i3x.FindFocused(tree.Root)
	if focused == nil {
		return nil
	}

	var target *i3.Node
	if !focused.IsFloating() {
		if workspace := i3x.WorkspaceOf(tree.Root, focused.ID); workspace != nil {
			target = neighbour(focused, i3x.Leaves(workspace), direction)
		}
	}

	if target == nil {
		_, err := i3.RunCommand("move " + direction)
		return err
	}

	if _, err := i3.RunCommand(fmt.Sprintf("swap container with con_id %d", target.ID)); err != nil {
		return err
	}
	_, err = i3.RunCommand(fmt.Sprintf("[con_id=%d] focus", focused.ID))
	return err
}

func neighbour(focused *i3.Node, leaves []*i3.Node, direction string) *i3.Node {
	var best *i3.Node
	var bestRank rank
	for _, leaf := range leaves {
		if leaf.ID == focused.ID {
			continue
		}
		current, ok := ranking(focused.Rect, leaf.Rect, direction)
		if !ok {
			continue
		}
		if best == nil || current.less(bestRank) {
			best, bestRank = leaf, current
		}
	}
	return best
}

type rank struct {
	edge   int64
	offset int64
}

func (r rank) less(other rank) bool {
	if r.edge != other.edge {
		return r.edge < other.edge
	}
	return r.offset < other.offset
}

func ranking(focused, candidate i3.Rect, direction string) (rank, bool) {
	switch direction {
	case "left":
		if candidate.X+candidate.Width > focused.X || !overlapsVertically(focused, candidate) {
			return rank{}, false
		}
		return rank{-(candidate.X + candidate.Width), distance(centerY(candidate), centerY(focused))}, true
	case "right":
		if candidate.X < focused.X+focused.Width || !overlapsVertically(focused, candidate) {
			return rank{}, false
		}
		return rank{candidate.X, distance(centerY(candidate), centerY(focused))}, true
	case "up":
		if candidate.Y+candidate.Height > focused.Y || !overlapsHorizontally(focused, candidate) {
			return rank{}, false
		}
		return rank{-(candidate.Y + candidate.Height), distance(centerX(candidate), centerX(focused))}, true
	default:
		if candidate.Y < focused.Y+focused.Height || !overlapsHorizontally(focused, candidate) {
			return rank{}, false
		}
		return rank{candidate.Y, distance(centerX(candidate), centerX(focused))}, true
	}
}

func overlapsVertically(a, b i3.Rect) bool {
	return a.Y < b.Y+b.Height && b.Y < a.Y+a.Height
}

func overlapsHorizontally(a, b i3.Rect) bool {
	return a.X < b.X+b.Width && b.X < a.X+a.Width
}

func centerX(r i3.Rect) int64 { return 2*r.X + r.Width }

func centerY(r i3.Rect) int64 { return 2*r.Y + r.Height }

func distance(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}
