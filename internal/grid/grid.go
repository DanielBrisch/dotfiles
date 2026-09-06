package grid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	i3 "go.i3wm.org/i3/v4"

	"github.com/DanielBrisch/dotfiles/internal/i3x"
)

const (
	TmpWorkspace   = "_gridtmp"
	TickFlip       = "i3grid:flip"
	internalPrefix = "__i3"
)

type Grid struct {
	busy      bool
	stackLast bool
}

func Once(name string) error {
	grid := &Grid{}
	tree, err := i3.GetTree()
	if err != nil {
		return err
	}
	if name == "" {
		workspace := i3x.FocusedWorkspace(tree.Root)
		if workspace == nil {
			return nil
		}
		name = workspace.Name
	}
	grid.rebuild(name, 0)
	return nil
}

func Rescue() error {
	tree, err := i3.GetTree()
	if err != nil {
		return err
	}
	workspace := i3x.FocusedWorkspace(tree.Root)
	if workspace == nil {
		return nil
	}
	(&Grid{}).rescue(workspace.Name)
	return nil
}

func Flip() error {
	_, err := i3.SendTick(TickFlip)
	return err
}

func Run() error {
	stopOtherInstances()

	grid := &Grid{}
	receiver := i3.Subscribe(i3.WindowEventType, i3.TickEventType)
	for receiver.Next() {
		switch event := receiver.Event().(type) {
		case *i3.WindowEvent:
			grid.onWindow(event)
		case *i3.TickEvent:
			grid.onTick(event)
		}
	}
	return receiver.Close()
}

func (g *Grid) onWindow(event *i3.WindowEvent) {
	switch event.Change {
	case "new":
		g.rebuildFocused(event.Container.ID)
	case "close":
		g.rebuildFocused(0)
	case "move", "floating":
		g.rebuildMoved(event.Container.ID)
	}
}

func (g *Grid) onTick(event *i3.TickEvent) {
	if event.Payload != TickFlip {
		return
	}
	g.stackLast = !g.stackLast
	side := "inicial"
	if g.stackLast {
		side = "final"
	}
	log.Printf("pilha na coluna %s", side)
	g.rebuildFocused(0)
}

func (g *Grid) rebuildFocused(preferFocus i3.NodeID) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Printf("arvore indisponivel: %v", err)
		return
	}
	workspace := i3x.FocusedWorkspace(tree.Root)
	if workspace == nil {
		return
	}
	g.rebuild(workspace.Name, preferFocus)
}

func (g *Grid) rebuildMoved(id i3.NodeID) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Printf("arvore indisponivel: %v", err)
		return
	}

	var names []string
	if workspace := i3x.WorkspaceOf(tree.Root, id); workspace != nil {
		names = append(names, workspace.Name)
	}
	if workspace := i3x.FocusedWorkspace(tree.Root); workspace != nil && !slices.Contains(names, workspace.Name) {
		names = append(names, workspace.Name)
	}

	for _, name := range names {
		if name == TmpWorkspace || strings.HasPrefix(name, internalPrefix) {
			continue
		}
		g.rebuild(name, 0)
	}
}

func (g *Grid) rebuild(name string, preferFocus i3.NodeID) {
	if g.busy {
		return
	}
	g.busy = true
	defer func() { g.busy = false }()

	tree, err := i3.GetTree()
	if err != nil {
		log.Printf("arvore indisponivel: %v", err)
		return
	}
	workspace := i3x.FindWorkspace(tree.Root, name)
	if workspace == nil || name == TmpWorkspace {
		return
	}

	g.rescue(name)

	tree, err = i3.GetTree()
	if err != nil {
		log.Printf("arvore indisponivel: %v", err)
		return
	}
	workspace = i3x.FindWorkspace(tree.Root, name)
	if workspace == nil {
		return
	}

	leaves := i3x.Leaves(workspace)
	if len(leaves) < 2 {
		return
	}
	windows := make([]i3.NodeID, 0, len(leaves))
	for _, leaf := range leaves {
		if leaf.ID != preferFocus {
			windows = append(windows, leaf.ID)
		}
	}
	if len(windows) != len(leaves) {
		windows = append(windows, preferFocus)
	}

	if sameGrid(Partition(windows, ColumnsFor(len(windows)), g.stackLast), CurrentGrid(workspace)) {
		return
	}

	before := Proportions(workspace)
	for _, id := range windows {
		run(fmt.Sprintf("[con_id=%d] move container to workspace \"%s\"", id, TmpWorkspace))
	}
	g.layOut(name, windows, preferFocus)
	g.rescue(name)
	g.restore(name, before)
}

func (g *Grid) rescue(target string) {
	tree, err := i3.GetTree()
	if err != nil {
		return
	}
	stranded := i3x.FindWorkspace(tree.Root, TmpWorkspace)
	if stranded == nil {
		return
	}
	for _, leaf := range i3x.Leaves(stranded) {
		log.Printf("resgatando janela %d para %s", leaf.ID, target)
		run(fmt.Sprintf("[con_id=%d] move container to workspace \"%s\"", leaf.ID, target))
	}
}

func (g *Grid) layOut(name string, wanted []i3.NodeID, preferFocus i3.NodeID) {
	tree, err := i3.GetTree()
	if err != nil {
		return
	}
	parked := i3x.FindWorkspace(tree.Root, TmpWorkspace)
	if parked == nil {
		return
	}

	available := make(map[i3.NodeID]bool)
	for _, leaf := range i3x.Leaves(parked) {
		available[leaf.ID] = true
	}
	windows := make([]i3.NodeID, 0, len(wanted))
	for _, id := range wanted {
		if available[id] {
			windows = append(windows, id)
		}
	}
	if len(windows) == 0 {
		return
	}

	var previous i3.NodeID
	for _, column := range Partition(windows, ColumnsFor(len(windows)), g.stackLast) {
		for position, id := range column {
			moved := fmt.Sprintf("[con_id=%d] move container to workspace \"%s\"", id, name)
			var ok bool
			switch {
			case previous == 0:
				ok = run(moved) && run(fmt.Sprintf("[con_id=%d] focus", id)) &&
					run("layout splith") && run("split v")
			case position == 0:
				ok = run(fmt.Sprintf("[con_id=%d] focus", previous)) && run(moved) &&
					run(fmt.Sprintf("[con_id=%d] move right", id)) &&
					run(fmt.Sprintf("[con_id=%d] focus", id)) && run("split v")
			default:
				ok = run(fmt.Sprintf("[con_id=%d] focus", previous)) && run(moved)
			}
			if ok {
				previous = id
			}
		}
	}

	target := windows[0]
	if slices.Contains(windows, preferFocus) {
		target = preferFocus
	}
	run(fmt.Sprintf("[con_id=%d] focus", target))
}

func (g *Grid) restore(name string, before []Column) {
	tree, err := i3.GetTree()
	if err != nil {
		return
	}
	workspace := i3x.FindWorkspace(tree.Root, name)
	if workspace == nil || len(workspace.Nodes) == 0 || len(workspace.Nodes) != len(before) {
		return
	}

	for index, node := range workspace.Nodes[:len(workspace.Nodes)-1] {
		run(fmt.Sprintf("[con_id=%d] resize set width %d ppt", node.ID, before[index].Width))
	}
	for index, node := range workspace.Nodes {
		heights := before[index].Heights
		leaves := i3x.Leaves(node)
		if node.Window != 0 || len(leaves) != len(heights) || len(leaves) < 2 {
			continue
		}
		for position, leaf := range leaves[:len(leaves)-1] {
			run(fmt.Sprintf("[con_id=%d] resize set height %d ppt", leaf.ID, heights[position]))
		}
	}
}

func run(command string) bool {
	results, err := i3.RunCommand(command)
	if err != nil && !i3.IsUnsuccessful(err) {
		log.Printf("comando falhou: %s -> %v", command, err)
		return false
	}
	for _, result := range results {
		if !result.Success {
			log.Printf("comando recusado: %s -> %s", command, result.Error)
			return false
		}
	}
	return true
}

func stopOtherInstances() {
	self, err := os.Executable()
	if err != nil {
		return
	}
	mine := os.Getpid()

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}
	for _, entry := range entries {
		pid := 0
		if _, err := fmt.Sscanf(entry.Name(), "%d", &pid); err != nil || pid == mine {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil {
			continue
		}
		parts := strings.Split(strings.TrimRight(string(raw), "\x00"), "\x00")
		if len(parts) < 2 || parts[0] != self || parts[1] != "daemon" {
			continue
		}
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.Printf("nao consegui encerrar o pid %d: %v", pid, err)
		}
	}
}
