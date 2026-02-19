package vcs

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

const (
	tempDir = "temp"
)

var (
	errNothingToCommit = errors.New("nothing to commit")
)

type commit struct {
	hash          hash.Hash
	workingDir    string
	commitDirPath string
}

func newCommit(workingDir string) (*commit, error) {
	if err := createCommitDirectory(temporaryCommitDirPath(workingDir)); err != nil {
		return nil, err
	}

	return &commit{
		hash:       sha256.New(),
		workingDir: workingDir,
	}, nil
}

func (c *commit) isEmpty() bool {
	return c.hash.Size() == 0
}

func (c *commit) copyFile(srcFilePath string) error {
	srcFile, err := c.readFile(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := c.hashFile(srcFile); err != nil {
		return err
	}

	if err := c.createFileDirectories(filepath.Join(temporaryCommitDirPath(c.workingDir), srcFilePath)); err != nil {
		return err
	}

	if err := c.resetReader(srcFile); err != nil {
		return err
	}

	dstFilePath := filepath.Join(temporaryCommitDirPath(c.workingDir), srcFilePath)
	w, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer w.Close()

	return copyFile(w, srcFile)
}

func (c *commit) execute() (string, error) {
	if c.isEmpty() {
		return "", errNothingToCommit
	}

	changesHash := fmt.Sprintf("%x", c.hash.Sum(nil))
	if err := os.Rename(temporaryCommitDirPath(c.workingDir), commitDirPath(c.workingDir, changesHash)); err != nil {
		return "", err
	}

	return changesHash, nil
}

func (c *commit) readFile(srcFilePath string) (*os.File, error) {
	return os.Open(filepath.Join(c.workingDir, srcFilePath))
}

func (c *commit) hashFile(file *os.File) error {
	_, err := io.Copy(c.hash, file)
	return err
}

func (c *commit) resetReader(file *os.File) error {
	_, err := file.Seek(0, io.SeekStart)
	return err
}

func (c *commit) createFileDirectories(srcFilePath string) error {
	return createDirectory(filepath.Dir(srcFilePath))
}

func createCommitDirectory(dirPath string) error {
	commitDirExist, err := fileExists(dirPath)
	if err != nil {
		return err
	}

	if !commitDirExist {
		return createDirectory(dirPath)
	}

	return nil
}

func temporaryCommitDirPath(workingDir string) string {
	return filepath.Join(workingDir, vcsDir, commitsDir, tempDir)
}

func commitDirPath(workingDir, hash string) string {
	return filepath.Join(workingDir, vcsDir, commitsDir, hash)
}
