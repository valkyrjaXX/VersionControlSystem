package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oc/vcs/internal/pkg/vcs"
)

type addCmd struct {
	commandMeta
}

func (c *addCmd) Run(ctx context.Context, vc *vcs.VersionControl, args ...string) string {
	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	default:
		workingRepository := vc.GetWorkingRepository()
		if workingRepository == nil {
			return fmt.Sprintf("Select working repository first. Available: %s", strings.Join(vc.ListRepositories(), ", "))
		}

		if len(args) == 0 {
			var indexContent string
			err := workingRepository.ReadIndex(func(data string) error {
				indexContent = data
				return nil
			})
			if err != nil {
				if errors.Is(err, vcs.ErrEmptyFile) {
					return "Add a file to the index."
				}

				return err.Error()
			}

			return fmt.Sprintf("Tracked files:\n%s", indexContent)
		}

		if err := workingRepository.WriteToIndex(args[0]); err != nil {
			return err.Error()
		}

		return fmt.Sprintf("The file '%s' is tracked.", args[0])
	}
}
