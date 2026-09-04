#!/usr/bin/env bash
set -euo pipefail

LOCAL_ENV="$HOME/.config/i3/local.env"
[ -f "$LOCAL_ENV" ] && . "$LOCAL_ENV"

CSS="${LOCK_ART_CSS:-}"
DIR="${XDG_CACHE_HOME:-$HOME/.cache}/i3"
NAME="${LOCK_NAME:-$(id -un)}"
PHRASE="${LOCK_PHRASE:-}"
LOCKER_COLOR="$HOME/.local/bin/i3lock-color"

mkdir -p "$DIR"

read -r W H < <(xrandr 2>/dev/null | grep -oE 'current [0-9]+ x [0-9]+' | head -1 | awk '{print $2, $4}') || true
: "${W:=1920}" "${H:=1080}"

mons=$(xrandr 2>/dev/null | awk '$2=="connected"{
  if ($3=="primary") { pri=1; geo=$4 } else { pri=0; geo=$3 }
  if (geo ~ /^[0-9]+x[0-9]+\+[0-9]+\+[0-9]+$/) print pri, geo
}')
primary=$(printf '%s\n' "$mons" | awk '$1==1{print $2; exit}')
[ -n "$primary" ] || primary=$(printf '%s\n' "$mons" | awk 'NF{print $2; exit}')
[ -n "$primary" ] || primary="${W}x${H}+0+0"
secondary=$(printf '%s\n' "$mons" | awk -v p="$primary" '$2!=p{print $2; exit}')
[ -n "$secondary" ] || secondary="$primary"

parse_geom() {
  local g=$1 rest
  GW=${g%%x*}; rest=${g#*x}
  GH=${rest%%+*}; rest=${rest#*+}
  GX=${rest%%+*}; GY=${rest#*+}
}

parse_geom "$primary";   PW=$GW PH=$GH PX=$GX PY=$GY
parse_geom "$secondary"; SW=$GW SH=$GH SX=$GX SY=$GY

sig=$(printf '%s|%s|%s' "$NAME" "$PHRASE" \
  "$(stat -c %Y "$CSS" 2>/dev/null || echo 0)" | md5sum | cut -c1-8)
layout=$(printf '%s|%s|%s' "${W}x${H}" "$primary" "$secondary" | md5sum | cut -c1-6)
OUT="$DIR/lockscreen-${W}x${H}-${layout}-${sig}.png"

printf '%s screen=%sx%s pri=%s sec=%s img=%s existe=%s\n' \
  "$(date -Is)" "$W" "$H" "$primary" "$secondary" "${OUT##*/}" \
  "$([ -f "$OUT" ] && echo sim || echo nao)" >> "$DIR/lock.log"

if [ "${1:-}" = "--debug" ]; then
  echo "screen (framebuffer): ${W}x${H}"
  echo "primary:   $primary"
  echo "secondary: $secondary"
  echo "cache:     $OUT"
  echo "existe:    $([ -f "$OUT" ] && echo sim || echo NAO)"
  echo "--- xrandr --listmonitors ---"; xrandr --listmonitors
  echo "--- outputs ---"; xrandr | grep -E ' connected| disconnected'
  echo "--- PNGs em cache ---"; ls -1 "$DIR"/lockscreen-*.png 2>/dev/null || echo "(nenhum)"
  exit 0
fi

if [ ! -f "$OUT" ]; then
  tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
  FONT=$(fc-match "JetBrains Mono:medium" -f "%{file}")
  FONT_JP=$(fc-match ":lang=ja" -f "%{file}")
  ART_HEIGHT=$(( SH * 55 / 100 ))
  ART_RIGHT=$(( W - SX - SW ))
  ART_BOTTOM=$(( H - SY - SH ))
  TEXT_LEFT=$(( PX + 80 ))
  TEXT_BOTTOM=$(( H - PY - PH + 80 ))

  art=()
  if [ -n "$CSS" ] && [ -r "$CSS" ] \
     && grep -qo 'data:image/[a-z]*;base64,[A-Za-z0-9+/=]*' "$CSS"; then
    grep -o 'data:image/[a-z]*;base64,[A-Za-z0-9+/=]*' "$CSS" | head -1 \
      | sed 's/^data:image\/[a-z]*;base64,//' | base64 -d > "$tmp/art.src"

    convert "$tmp/art.src" -colorspace Gray -negate \
      -function Polynomial 1.7,-0.35 -resize "x${ART_HEIGHT}" "$tmp/art.png"

    read -r AW AH < <(identify -format "%w %h\n" "$tmp/art.png")

    convert -size "${AW}x${AH}" xc:black \
      -fx "dd=(((w-1-i)/w)+((h-1-j)/h))/2; dd<=0.3?1:(dd>=0.9?0:(0.9-dd)/0.6)" \
      "$tmp/mask.png"

    convert "$tmp/art.png" "$tmp/mask.png" -alpha off -compose CopyOpacity -composite \
      -channel A -evaluate multiply 0.6 +channel "$tmp/art-final.png"

    art=("$tmp/art-final.png" -gravity SouthEast
         -geometry "+${ART_RIGHT}+${ART_BOTTOM}" -composite)
  fi

  convert -size "${W}x${H}" xc:black \
    ${art[@]+"${art[@]}"} \
    -gravity SouthWest \
    -font "$FONT_JP" -pointsize 72 -fill '#888888' \
    -annotate "+${TEXT_LEFT}+$(( TEXT_BOTTOM + 90 ))" "$PHRASE" \
    -font "$FONT" -pointsize 56 -fill white \
    -annotate "+${TEXT_LEFT}+${TEXT_BOTTOM}" "$NAME" \
    -depth 8 -colorspace sRGB -alpha off "PNG24:$OUT"

  find "$DIR" -maxdepth 1 -name 'lockscreen-*.png' ! -name "*-${sig}.png" -delete
fi

if [ -x "$LOCKER_COLOR" ]; then
  screen_idx=$(xrandr --listmonitors 2>/dev/null \
    | awk -v g="$secondary" 'NR>1{gsub("/[0-9]*","",$3); if ($3==g) {print NR-1; exit}}')
  [ -n "$screen_idx" ] || screen_idx=0
  exec "$LOCKER_COLOR" -i "$OUT" -c 000000 -k \
    --screen "$screen_idx" \
    --time-str="%H:%M:%S" --date-str="" \
    --time-font="JetBrains Mono" --time-size=96 --time-color=ffffffff \
    --time-pos="x+w/2:y+h/2"
fi

exec i3lock -c 000000 -i "$OUT"
