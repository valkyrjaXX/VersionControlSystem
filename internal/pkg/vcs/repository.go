package vcs

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	vcsDirPath = "./vcs"

	commitsDir = "commits"

	indexFile = "index.txt"
	logFile   = "log.txt"

	comma = ","
)

var (
	readWriteFileMode = os.FileMode(0444)
)

type Repository struct {
	path string
}

func newRepository(rootDir string, dir string) (*Repository, error) {
	dirPath := filepath.Join(rootDir, dir)
	_, err := os.Stat(dirPath)
	if err == nil {
		return &Repository{path: dirPath}, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		err = createRepoStructure(dirPath)
		if err != nil {
			return nil, err
		}

		return &Repository{
			path: dirPath,
		}, nil
	}

	return nil, err
}

func createRepoStructure(dirPath string) error {
	err := os.MkdirAll(filepath.Join(dirPath, commitsDir), readWriteFileMode)
	if err != nil {
		return err
	}

	if err = createFile(filepath.Join(dirPath, configFile)); err != nil {
		return err
	}

	if err = createFile(filepath.Join(dirPath, indexFile)); err != nil {
		return err
	}

	if err = createFile(filepath.Join(dirPath, logFile)); err != nil {
		return err
	}

	return nil
}
