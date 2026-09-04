#!/usr/bin/env bash
set -euo pipefail

SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/config"
DEST="${XDG_CONFIG_HOME:-$HOME/.config}"
STAMP="$(date +%Y%m%d-%H%M%S)"

link() {
    local src="$SRC/$1" dst="$DEST/$1"
    mkdir -p "$(dirname "$dst")"
    if [ -L "$dst" ]; then
        rm "$dst"
    elif [ -e "$dst" ]; then
        mv "$dst" "$dst.bak-$STAMP"
        echo "backup:  $dst.bak-$STAMP"
    fi
    ln -s "$src" "$dst"
    echo "link:    $dst -> $src"
}

for item in i3 i3blocks kitty rofi picom.conf; do
    link "$item"
done

for f in local.conf local.env; do
    [ -e "$SRC/i3/$f" ] || { cp "$SRC/i3/$f.example" "$SRC/i3/$f"; echo "criado:  $SRC/i3/$f"; }
done

echo
echo "Pronto. Ajuste $SRC/i3/local.env e recarregue o i3 com \$mod+Shift+c."
