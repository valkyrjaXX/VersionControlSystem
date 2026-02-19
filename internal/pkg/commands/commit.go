package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type commitCmd struct {
	commandMeta
}

func (c *commitCmd) Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string {
	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	default:
		workingRepository := vc.GetWorkingRepository()
		if workingRepository == nil {
			return fmt.Sprintf("Select working repository first. Available: %s", strings.Join(vc.ListRepositories(), ", "))
		}

		if len(args) == 0 {
			return "Message was not passed."
		}

		message := args[0]
		username, err := vc.ReadConfig()
		if err != nil {
			return err.Error()
		}

		hash, err := workingRepository.Commit(username, message)
		if err != nil {
			if errors.Is(err, vcs.ErrNothingToCommit) {
				return "Nothing to commit."
			}

			return err.Error()
		}

		return fmt.Sprintf("Changes are committed. %s", hash)
	}
}
