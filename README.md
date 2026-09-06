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

## i3grid

Os scripts de layout são um binário Go em `cmd/i3grid`, compilado pelo
`install.sh` para `~/.local/bin/i3grid`.

| comando | onde é usado |
| --- | --- |
| `i3grid daemon` | `exec_always` do i3; mantém a grade a cada janela aberta, fechada ou movida |
| `i3grid move <dir>` | `$mod+Shift+<direção>`; troca a janela focada com a vizinha real daquele lado |
| `i3grid flip` | `$mod+g`; inverte de que lado fica a coluna empilhada |
| `i3grid once [-workspace <nome>]` | reorganiza uma vez, sem daemon |
| `i3grid rescue` | traz de volta janelas presas no workspace temporário |

```sh
go test ./...
```

## Dependências

`i3`, `i3blocks`, `rofi`, `picom`, `kitty`, `feh`, `xss-lock`, `i3lock`,
`imagemagick`, `flameshot`, `copyq`, `brightnessctl`, `redshift`, `dex`,
`jq` (usado pelo `swap.sh`) e a toolchain do `go` para compilar o `i3grid`.

Opcional: `i3lock-color` em `~/.local/bin/i3lock-color` — sem ele o `lock.sh`
cai no `i3lock` normal.
