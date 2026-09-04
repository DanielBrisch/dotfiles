#!/usr/bin/env bash
set -u

LOCAL_ENV="$HOME/.config/i3/local.env"
[[ -f $LOCAL_ENV ]] && . "$LOCAL_ENV"

WALLPAPER="${WALLPAPER:-}"
PIDFILE="${XDG_RUNTIME_DIR:-/tmp}/i3-wallpaper-daemon.pid"

if [[ -f $PIDFILE ]]; then
    old_pid=$(<"$PIDFILE")
    if [[ $old_pid != "$$" ]] && kill -0 "$old_pid" 2>/dev/null; then
        kill "$old_pid" 2>/dev/null
        sleep 0.3
    fi
fi
echo $$ >"$PIDFILE"

xev_pid=
debounce_pid=

cleanup() {
    [[ -n $xev_pid ]] && kill "$xev_pid" 2>/dev/null
    [[ -n $debounce_pid ]] && kill "$debounce_pid" 2>/dev/null
    [[ -f $PIDFILE && $(<"$PIDFILE") == "$$" ]] && rm -f "$PIDFILE"
    return 0
}
trap cleanup EXIT INT TERM

apply_wallpaper() {
    [[ -r $WALLPAPER ]] || return 0
    feh --no-fehbg --bg-scale "$WALLPAPER"
}

apply_wallpaper

exec 3< <(stdbuf -oL xev -root -event randr)
xev_pid=$!

while read -r -u 3 line; do
    if [[ $line == RRScreenChangeNotify* ]]; then
        [[ -n $debounce_pid ]] && kill "$debounce_pid" 2>/dev/null
        {
            sleep 1
            apply_wallpaper
        } &
        debounce_pid=$!
    fi
done
