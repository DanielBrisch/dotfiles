#!/usr/bin/env bash
set -euo pipefail

MARK="swap"

focused_is_marked() {
    i3-msg -t get_tree \
        | jq -e --arg m "$MARK" \
            'first(.. | objects | select(.focused == true) | .marks // []) | index($m)' \
        >/dev/null
}

mark_exists() {
    i3-msg -t get_marks | jq -e --arg m "$MARK" 'index($m)' >/dev/null
}

if focused_is_marked; then
    i3-msg "unmark \"$MARK\""
elif mark_exists; then
    i3-msg "swap container with mark \"$MARK\"; unmark \"$MARK\""
else
    i3-msg "mark --add \"$MARK\""
fi
