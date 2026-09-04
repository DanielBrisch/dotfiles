#!/usr/bin/env python3
import sys

import i3ipc

DIRECTIONS = ("left", "right", "up", "down")


def center_x(rect):
    return rect.x + rect.width / 2


def center_y(rect):
    return rect.y + rect.height / 2


def overlaps_vertically(a, b):
    return a.y < b.y + b.height and b.y < a.y + a.height


def overlaps_horizontally(a, b):
    return a.x < b.x + b.width and b.x < a.x + a.width


def ranking(focused, candidate, direction):
    a, b = focused.rect, candidate.rect
    if direction == "left":
        if b.x + b.width > a.x or not overlaps_vertically(a, b):
            return None
        return (-(b.x + b.width), abs(center_y(b) - center_y(a)))
    if direction == "right":
        if b.x < a.x + a.width or not overlaps_vertically(a, b):
            return None
        return (b.x, abs(center_y(b) - center_y(a)))
    if direction == "up":
        if b.y + b.height > a.y or not overlaps_horizontally(a, b):
            return None
        return (-(b.y + b.height), abs(center_x(b) - center_x(a)))
    if b.y < a.y + a.height or not overlaps_horizontally(a, b):
        return None
    return (b.y, abs(center_x(b) - center_x(a)))


def neighbour(focused, leaves, direction):
    best = None
    best_rank = None
    for leaf in leaves:
        if leaf.id == focused.id:
            continue
        rank = ranking(focused, leaf, direction)
        if rank is None:
            continue
        if best_rank is None or rank < best_rank:
            best, best_rank = leaf, rank
    return best


def main():
    if len(sys.argv) != 2 or sys.argv[1] not in DIRECTIONS:
        print("uso: move-swap.py <left|right|up|down>", file=sys.stderr)
        return 2

    direction = sys.argv[1]
    i3 = i3ipc.Connection()
    focused = i3.get_tree().find_focused()
    if focused is None:
        return 0

    workspace = focused.workspace()
    floating = focused.floating in ("user_on", "auto_on")

    target = None
    if workspace is not None and not floating:
        target = neighbour(focused, workspace.leaves(), direction)

    if target is None:
        i3.command("move %s" % direction)
        return 0

    i3.command("swap container with con_id %d" % target.id)
    i3.command("[con_id=%d] focus" % focused.id)
    return 0


if __name__ == "__main__":
    sys.exit(main())
