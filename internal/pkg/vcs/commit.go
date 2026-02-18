package vcs

import (
	"hash"
	"io"
	"os"
	"path/filepath"
)

type Commit struct {
	Uuid          string
	Hash          hash.Hash
	WorkingDir    string
	CommitDirPath string
	Username      string
	Message       string
}

func (c *Commit) CopyFile(srcFilePath string) error {
	if err := c.createCommitDirectory(); err != nil {
		return err
	}

	srcFile, err := c.readFile(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := c.hashFile(srcFile); err != nil {
		return err
	}

	if err := c.createFileDirectories(srcFilePath); err != nil {
		return err
	}

	if err := c.resetReader(srcFile); err != nil {
		return err
	}

	dstFilePath := filepath.Join(c.CommitDirPath, srcFilePath)
	w, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer w.Close()

	return c.copyFile(w, srcFile)
}

func (c *Commit) createCommitDirectory() error {
	commitDirExist, err := fileExists(c.CommitDirPath)
	if err != nil {
		return err
	}

	if !commitDirExist {
		return createDirectory(c.CommitDirPath)
	}

	return nil
}

func (c *Commit) readFile(srcFilePath string) (*os.File, error) {
	return os.Open(filepath.Join(c.WorkingDir, srcFilePath))
}

func (c *Commit) hashFile(file *os.File) error {
	_, err := io.Copy(c.Hash, file)
	return err
}

func (c *Commit) resetReader(file *os.File) error {
	_, err := file.Seek(0, io.SeekStart)
	return err
}

func (c *Commit) copyFile(dstFile *os.File, srcFile *os.File) error {
	_, err := io.Copy(dstFile, srcFile)
	return err
}

func (c *Commit) createFileDirectories(srcFilePath string) error {
	parentDir := filepath.Dir(srcFilePath)
	return createDirectory(filepath.Join(c.CommitDirPath, parentDir))
}
