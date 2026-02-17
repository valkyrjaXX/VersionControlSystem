package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/oc/vcs/internal/pkg/actions"
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

	cli := actions.NewCli(versionControlSystem)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		fmt.Println(cli.Command(parts[0], parts[1:]...))
	}
}
