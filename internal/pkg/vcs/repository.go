package vcs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	vcsDir     = ".vcs"
	commitsDir = "commits"

	indexFile = "index.txt"
	logFile   = "log.txt"

	comma = ","
)

var (
	readWriteFileMode = os.FileMode(0444)

	ErrEmptyFile       = errors.New("empty file")
	ErrFileNotExist    = errors.New("file does not exist")
	ErrNothingToCommit = errors.New("nothing to commit")
	ErrCommitNotFound  = errors.New("commit not found")
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

func (r *Repository) ClearIndex() error {
	filePath := filepath.Join(r.path, vcsDir, indexFile)
	return os.Truncate(filePath, 0)
}

func (r *Repository) Commit(username string, message string) (string, error) {
	c, err := newCommit(r.path)
	if err != nil {
		return "", err
	}

	if err = r.ReadIndex(func(filePath string) error {
		return c.copyFile(filePath)
	}); err != nil {
		return "", err
	}

	hash, err := c.execute()
	if err != nil {
		return "", err
	}

	if err := r.writeLog(hash, username, message); err != nil {
		return "", err
	}

	return hash, r.ClearIndex()
}

func (r *Repository) ReadLog(callback func(commit, author, comment string)) error {
	filePath := filepath.Join(r.path, vcsDir, logFile)
	read, err := readFileContent(filePath, func(data string) error {
		dataSlice := strings.SplitN(data, comma, 3)
		callback(dataSlice[0], dataSlice[1], dataSlice[2])
		return nil
	})
	if err != nil {
		return err
	}

	if read == 0 {
		return ErrEmptyFile
	}

	return nil
}

func (r *Repository) Checkout(commitID string) error {
	commitDirPath := filepath.Join(r.path, vcsDir, commitsDir, commitID)
	_, err := os.Stat(commitDirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrCommitNotFound
		}
		return err
	}

	return filepath.WalkDir(commitDirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		reader, err := os.Open(path)
		if err != nil {
			return err
		}
		defer reader.Close()

		dstFilePath := strings.ReplaceAll(path, commitDirPath, r.path)
		writer, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_TRUNC, readWriteFileMode)
		if err != nil {
			return err
		}

		defer writer.Close()

		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *Repository) writeLog(hash, username, message string) error {
	filePath := filepath.Join(r.path, vcsDir, logFile)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, readWriteFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	data := strings.Join([]string{hash, username, message}, comma)
	_, err = fmt.Fprintln(file, data)
	if err != nil {
		return err
	}

	return nil
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
