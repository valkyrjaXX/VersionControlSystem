package commands

import (
	"context"
	"fmt"

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
		if len(args) == 0 {
			return ""
		}

		err := vc.Checkout(args[0])
		if err != nil {
			return err.Error()
		}

		return fmt.Sprintf("Working directory %s.", vc.GetWorkingRepository())
	}
}
