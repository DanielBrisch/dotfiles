package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/DanielBrisch/dotfiles/internal/grid"
	"github.com/DanielBrisch/dotfiles/internal/move"
)

const usage = `uso: i3grid <comando>

  daemon                   escuta o i3 e mantem a grade
  once [-workspace <nome>]  reorganiza uma vez
  flip                     inverte de que lado fica a pilha
  rescue                   traz de volta janelas presas no workspace temporario
  move <left|right|up|down> troca a janela focada com a vizinha`

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(usage)
	}

	switch args[0] {
	case "daemon":
		return grid.Run()
	case "flip":
		return grid.Flip()
	case "rescue":
		return grid.Rescue()
	case "once":
		flags := flag.NewFlagSet("once", flag.ContinueOnError)
		workspace := flags.String("workspace", "", "nome do workspace (vazio = o focado)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return grid.Once(*workspace)
	case "move":
		if len(args) < 2 {
			return errors.New(usage)
		}
		return move.Run(args[1])
	}
	return fmt.Errorf("comando desconhecido: %s\n\n%s", args[0], usage)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("i3grid: ")
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
