#!/usr/bin/env python3
import math
import os
import signal
import sys
import traceback

import i3ipc

TMP_WORKSPACE = "_gridtmp"
SCRIPT = os.path.basename(__file__)


def stop_other_instances():
    mine = os.getpid()
    for entry in os.listdir("/proc"):
        if not entry.isdigit() or int(entry) == mine:
            continue
        try:
            with open("/proc/%s/cmdline" % entry, "rb") as handle:
                parts = handle.read().split(b"\0")
        except OSError:
            continue
        if not parts or b"python" not in os.path.basename(parts[0]):
            continue
        if any(SCRIPT.encode() in part for part in parts[1:]):
            try:
                os.kill(int(entry), signal.SIGTERM)
            except OSError:
                pass


def log(message):
    print(message, file=sys.stderr, flush=True)


def columns_for(count):
    return max(1, math.ceil(math.sqrt(count)))


def partition(windows, columns):
    result = []
    start = 0
    remaining = len(windows)
    for column in range(columns):
        take = math.ceil(remaining / (columns - column))
        result.append(windows[start:start + take])
        start += take
        remaining -= take
    return [column for column in result if column]


def current_grid(workspace):
    grid = []
    for node in workspace.nodes:
        grid.append([node.id] if node.window else [leaf.id for leaf in node.leaves()])
    return grid


def find_workspace(tree, name):
    return next((w for w in tree.workspaces() if w.name == name), None)


class Grid:
    def __init__(self, connection):
        self.i3 = connection
        self.busy = False

    def run(self, command):
        try:
            replies = self.i3.command(command)
        except Exception as error:
            log("comando falhou: %s -> %s" % (command, error))
            return False
        for reply in replies:
            if not reply.success:
                log("comando recusado: %s -> %s" % (command, reply.error))
                return False
        return True

    def rescue(self, target):
        tree = self.i3.get_tree()
        stranded = find_workspace(tree, TMP_WORKSPACE)
        if stranded is None:
            return
        for leaf in stranded.leaves():
            log("resgatando janela %d para %s" % (leaf.id, target))
            self.run('[con_id=%d] move container to workspace "%s"' % (leaf.id, target))

    def rebuild(self, workspace_name=None, prefer_focus=None):
        if self.busy:
            return
        self.busy = True
        try:
            self._rebuild(workspace_name, prefer_focus)
        except Exception:
            log(traceback.format_exc())
        finally:
            self.busy = False

    def _rebuild(self, workspace_name, prefer_focus):
        tree = self.i3.get_tree()
        if workspace_name is None:
            focused = tree.find_focused()
            workspace = focused.workspace() if focused else None
        else:
            workspace = find_workspace(tree, workspace_name)
        if workspace is None or workspace.name == TMP_WORKSPACE:
            return

        name = workspace.name
        self.rescue(name)

        workspace = find_workspace(self.i3.get_tree(), name)
        if workspace is None:
            return

        windows = [leaf.id for leaf in workspace.leaves()]
        if len(windows) < 2:
            return
        if prefer_focus in windows:
            windows.remove(prefer_focus)
            windows.append(prefer_focus)
        if partition(windows, columns_for(len(windows))) == current_grid(workspace):
            return

        try:
            for con_id in windows:
                self.run('[con_id=%d] move container to workspace "%s"' % (con_id, TMP_WORKSPACE))
            self._lay_out(name, windows, prefer_focus)
        finally:
            self.rescue(name)

    def _lay_out(self, name, wanted, prefer_focus):
        parked = find_workspace(self.i3.get_tree(), TMP_WORKSPACE)
        if parked is None:
            return
        available = {leaf.id for leaf in parked.leaves()}
        windows = [con_id for con_id in wanted if con_id in available]
        if not windows:
            return

        previous = None
        for column in partition(windows, columns_for(len(windows))):
            for position, con_id in enumerate(column):
                moved = '[con_id=%d] move container to workspace "%s"' % (con_id, name)
                if previous is None:
                    ok = self.run(moved) and self.run("[con_id=%d] focus" % con_id)
                    ok = ok and self.run("layout splith") and self.run("split v")
                elif position == 0:
                    ok = self.run("[con_id=%d] focus" % previous) and self.run(moved)
                    ok = ok and self.run("[con_id=%d] move right" % con_id)
                    ok = ok and self.run("[con_id=%d] focus" % con_id) and self.run("split v")
                else:
                    ok = self.run("[con_id=%d] focus" % previous) and self.run(moved)
                if ok:
                    previous = con_id

        target = prefer_focus if prefer_focus in windows else windows[0]
        self.run("[con_id=%d] focus" % target)


def main():
    arguments = sys.argv[1:]
    i3 = i3ipc.Connection()
    grid = Grid(i3)

    if "--rescue" in arguments:
        focused = i3.get_tree().find_focused()
        workspace = focused.workspace() if focused else None
        if workspace is not None:
            grid.rescue(workspace.name)
        return 0

    if "--once" in arguments:
        name = None
        if "--workspace" in arguments:
            index = arguments.index("--workspace")
            if index + 1 < len(arguments):
                name = arguments[index + 1]
        grid.rebuild(workspace_name=name)
        return 0

    def on_new(_, event):
        container = getattr(event, "container", None)
        grid.rebuild(prefer_focus=container.id if container else None)

    def on_close(_, __):
        grid.rebuild()

    stop_other_instances()
    i3.on(i3ipc.Event.WINDOW_NEW, on_new)
    i3.on(i3ipc.Event.WINDOW_CLOSE, on_close)
    i3.main()
    return 0


if __name__ == "__main__":
    sys.exit(main())
