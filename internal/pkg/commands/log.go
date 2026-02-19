package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type logCmd struct {
	commandMeta
}

func (c *logCmd) Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string {
	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	default:
		workingRepository := vc.GetWorkingRepository()
		if workingRepository == nil {
			return fmt.Sprintf("Select working repository first. Available: %s", strings.Join(vc.ListRepositories(), ", "))
		}

		var result []string
		err := workingRepository.ReadLog(func(commit, author, comment string) {
			defer func() {
				result = append(result, fmt.Sprintf("commit %s, author: %s, comment: %s", commit, author, comment))
			}()
		})
		if err != nil {
			if errors.Is(err, vcs.ErrEmptyFile) {
				return "No commits yet."
			}

			return err.Error()
		}

		return strings.Join(result, "\n")
	}
}
