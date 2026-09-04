#!/bin/bash

# Detecta os monitores conectados
monitors=$(xrandr | grep " connected" | awk '{print $1}')
monitor_array=($monitors)

# Se houver apenas um monitor, avisa
if [ ${#monitor_array[@]} -lt 2 ]; then
    notify-send "Monitor" "Apenas um monitor detectado"
    exit 0
fi

primary="${monitor_array[0]}"
secondary="${monitor_array[1]}"

# Menu de opções
options="Espelho (Mirror)\nEstender à Direita\nEstender à Esquerda\nEstender Acima\nEstender Abaixo\nApenas Monitor Principal\nApenas Monitor Secundário"

chosen=$(echo -e "$options" | rofi -dmenu -i -p "Configuração de Monitores" -theme-str 'window {width: 400px;}')

case "$chosen" in
    "Espelho (Mirror)")
        xrandr --output $primary --auto --output $secondary --auto --same-as $primary
        notify-send "Monitores" "Modo espelho ativado"
        ;;
    "Estender à Direita")
        xrandr --output $primary --auto --output $secondary --auto --right-of $primary
        notify-send "Monitores" "Monitor secundário à direita"
        ;;
    "Estender à Esquerda")
        xrandr --output $primary --auto --output $secondary --auto --left-of $primary
        notify-send "Monitores" "Monitor secundário à esquerda"
        ;;
    "Estender Acima")
        xrandr --output $primary --auto --output $secondary --auto --above $primary
        notify-send "Monitores" "Monitor secundário acima"
        ;;
    "Estender Abaixo")
        xrandr --output $primary --auto --output $secondary --auto --below $primary
        notify-send "Monitores" "Monitor secundário abaixo"
        ;;
    "Apenas Monitor Principal")
        xrandr --output $primary --auto --output $secondary --off
        notify-send "Monitores" "Usando apenas monitor principal"
        ;;
    "Apenas Monitor Secundário")
        xrandr --output $primary --off --output $secondary --auto
        notify-send "Monitores" "Usando apenas monitor secundário"
        ;;
esac
