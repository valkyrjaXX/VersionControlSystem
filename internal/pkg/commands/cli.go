package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type commandMeta struct {
	name        string
	description string
}

func (m commandMeta) Name() string        { return m.name }
func (m commandMeta) Description() string { return m.description }

type Command interface {
	Name() string
	Description() string
	Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string
}

type VcsCli interface {
	Help() string
	Exit() string
	Command(ctx context.Context, name string, args ...string) string
}

type cli struct {
	commands []Command
	vc       *vcs.VersionControl
}

func NewCli(vc *vcs.VersionControl) VcsCli {
	return &cli{
		vc: vc,
		commands: []Command{
			&configCmd{commandMeta: commandMeta{name: "config", description: "Get and set a username."}},
			&repositoryCmd{commandMeta: commandMeta{name: "repository", description: "Checkout a repository."}},
			&addCmd{commandMeta: commandMeta{name: "add", description: "Add a file to the index."}},
			&commitCmd{commandMeta: commandMeta{name: "commit", description: "Save changes."}},
			&logCmd{commandMeta: commandMeta{name: "log", description: "Show commit logs."}},
			&checkoutCmd{commandMeta: commandMeta{name: "checkout", description: "Restore a file."}},
		},
	}
}

func (cli *cli) Help() string {
	text := []string{"These are SVC commands:"}

	for _, cmd := range cli.commands {
		text = append(text, fmt.Sprintf("%s: %s", cmd.Name(), cmd.Description()))
	}

	return strings.Join(text, "\n")
}

func (cli *cli) Exit() string {
	return "Bye!"
}

func (cli *cli) Command(ctx context.Context, name string, args ...string) string {
	for _, cmd := range cli.commands {
		if cmd.Name() == name {
			return cmd.Run(ctx, cli.vc, args...)
		}
	}

	return cli.Help()
}
