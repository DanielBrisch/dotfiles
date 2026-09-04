# dotfiles

Ambiente gráfico: i3 + i3blocks + rofi + picom + kitty.

## Instalação

```sh
git clone <url> ~/dotfiles
~/dotfiles/install.sh
```

`install.sh` cria symlinks em `~/.config` e move para `.bak-<timestamp>`
qualquer config que já exista.

## Configuração local

Nada específico de máquina fica versionado. Dois arquivos, ambos gerados a
partir dos `.example` pelo `install.sh`:

| arquivo | usado por | serve para |
| --- | --- | --- |
| `config/i3/local.conf` | `include` no fim do `config` do i3 | redshift, `workspace ... output`, binds da máquina |
| `config/i3/local.env` | `lock.sh`, `wallpaper-daemon.sh` | nome na tela de lock, wallpaper, arte do lock |

## Dependências

`i3`, `i3blocks`, `rofi`, `picom`, `kitty`, `feh`, `xss-lock`, `i3lock`,
`imagemagick`, `flameshot`, `copyq`, `brightnessctl`, `redshift`, `dex`,
`python3` (scripts de layout, via `i3ipc`).

Opcional: `i3lock-color` em `~/.local/bin/i3lock-color` — sem ele o `lock.sh`
cai no `i3lock` normal.
