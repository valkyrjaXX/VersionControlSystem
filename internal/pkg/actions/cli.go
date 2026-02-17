package actions

import (
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type CommandDef struct {
	name        string
	description string
}

type VcsCli interface {
	Help() string
	Command(name string, args ...string) string
}

type cli struct {
	commands []CommandDef
	vc       *vcs.VersionControl
}

func NewCli(vc *vcs.VersionControl) VcsCli {
	return &cli{
		vc: vc,
		commands: []CommandDef{
			{
				name:        "config",
				description: "Get and set a username.",
			},
			{
				name:        "switch",
				description: "Switch repository.",
			},
		},
	}
}

func (cli *cli) Help() string {
	text := []string{"These are SVC commands:"}

	for _, cmdDef := range cli.commands {
		text = append(text, fmt.Sprintf("%s: %s", cmdDef.name, cmdDef.description))
	}

	return strings.Join(text, "\n")
}

func (cli *cli) Command(name string, args ...string) string {
	switch name {
	case "config":
		return cli.config(args...)
	case "switch":
		return cli.switchRepo(args...)
	default:
		return cli.Help()
	}
}

func (cli *cli) config(args ...string) string {
	if len(args) == 0 {
		config, err := cli.vc.ReadConfig()
		if err != nil {
			return err.Error()
		}

		if len(config) == 0 {
			return "Please, tell me who you are."
		}

		return fmt.Sprintf("The username is %s.", strings.Trim(config, "\n"))
	}

	username := args[0]
	if err := cli.vc.WriteConfig(username); err != nil {
		return err.Error()
	}

	return fmt.Sprintf("The username is %s.", username)
}

func (cli *cli) switchRepo(args ...string) string {
	if len(args) == 0 {
		return ""
	}

	err := cli.vc.Switch(args[0])
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Switched to %s.", args[0])
}
