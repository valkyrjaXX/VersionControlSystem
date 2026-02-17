package vcs

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
)

func createDirectory(dirPath string) error {
	_, err := os.Stat(dirPath)
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(dirPath, fs.ModeDir)
	}

	return err
}

func createFile(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	return file.Close()
}

func readFileContent(filePath string, callback func(data string) error) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var read int
	for scanner.Scan() {
		text := scanner.Text()
		if err := callback(text); err != nil {
			return read, err
		}

		read += len([]byte(text))
	}

	return read, scanner.Err()
}
