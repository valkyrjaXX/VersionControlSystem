package vcs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	vcsDir     = "vcs"
	commitsDir = "commits"

	indexFile = "index.txt"
	logFile   = "log.txt"

	comma = ","
)

var (
	readWriteFileMode = os.FileMode(0444)

	ErrEmptyFile    = errors.New("empty file")
	ErrFileNotExist = errors.New("file does not exist")
)

type Repository struct {
	name string
	path string
}

func (r *Repository) GetName() string {
	return r.name
}

func (r *Repository) ReadIndex(callback func(data string) error) error {
	filePath := filepath.Join(r.path, vcsDir, indexFile)
	read, err := readFileContent(filePath, callback)
	if err != nil {
		return err
	}

	if read == 0 {
		return ErrEmptyFile
	}

	return nil
}

func (r *Repository) WriteToIndex(fileToAdd string) error {
	fileExist, err := fileExists(fmt.Sprintf("%s/%s", r.path, fileToAdd))
	if err != nil {
		return err
	}

	if !fileExist {
		return ErrFileNotExist
	}

	filePath := filepath.Join(r.path, vcsDir, indexFile)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, readWriteFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, fileToAdd)
	return err
}

func newRepository(vcsDir string, dir string) (*Repository, error) {
	dirPath := filepath.Join(vcsDir, dir)
	_, err := os.Stat(dirPath)
	if err == nil {
		return &Repository{name: dir, path: dirPath}, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		err = createRepoStructure(dirPath)
		if err != nil {
			return nil, err
		}

		return &Repository{
			name: dir,
			path: dirPath,
		}, nil
	}

	return nil, err
}

func createRepoStructure(dirPath string) error {
	vscDir := filepath.Join(dirPath, vcsDir)

	err := os.MkdirAll(filepath.Join(vscDir, commitsDir), readWriteFileMode)
	if err != nil {
		return err
	}

	if err = createFile(filepath.Join(vscDir, configFile)); err != nil {
		return err
	}

	if err = createFile(filepath.Join(vscDir, indexFile)); err != nil {
		return err
	}

	if err = createFile(filepath.Join(vscDir, logFile)); err != nil {
		return err
	}

	return nil
}
