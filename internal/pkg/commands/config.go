package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type configCmd struct {
	commandMeta
}

func (c *configCmd) Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string {
	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	default:
		if len(args) == 0 {
			config, err := vc.ReadConfig()
			if err != nil {
				return err.Error()
			}

			if len(config) == 0 {
				return "Please, tell me who you are."
			}

			return fmt.Sprintf("The username is %s.", strings.Trim(config, "\n"))
		}

		username := args[0]
		if err := vc.WriteConfig(username); err != nil {
			return err.Error()
		}

		return fmt.Sprintf("The username is %s.", username)
	}
}
