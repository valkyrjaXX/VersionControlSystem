package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/oc/vcs/internal/pkg/commands"
	"github.com/oc/vcs/internal/pkg/vcs"
)

const (
	vcsDirPath = "./vcs"
)

func main() {
	var vcsDir string
	flag.StringVar(&vcsDir, "vcsDir", vcsDirPath, "")
	flag.Parse()

	versionControlSystem, err := vcs.NewVersionControl(vcsDir)
	if err != nil {
		log.Fatal(err)
	}

	cli := commands.NewCli(versionControlSystem)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			fmt.Println(cli.Exit())
			return
		default:
			fmt.Print("> ")

			if !scanner.Scan() {
				break
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if parts[0] == "exit" {
				cancel()
				continue
			}

			fmt.Println(cli.Command(ctx, parts[0], parts[1:]...))
		}
	}
}
