package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type checkoutCmd struct {
	commandMeta
}

func (c *checkoutCmd) Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string {
	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	default:
		workingRepository := vc.GetWorkingRepository()
		if workingRepository == nil {
			return fmt.Sprintf("Select working repository first. Available: %s", strings.Join(vc.ListRepositories(), ", "))
		}

		if len(args) == 0 {
			return "Commit id was not passed."
		}

		commit := args[0]
		err := workingRepository.Checkout(commit)
		if err != nil {
			if errors.Is(err, vcs.ErrCommitNotFound) {
				return "Commit does not exist."
			}
			return err.Error()
		}

		return fmt.Sprintf("Switched to commit %s.", commit)
	}
}
