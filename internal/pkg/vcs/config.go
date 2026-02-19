package vcs

import (
	"fmt"
	"os"
	"sync"
)

type config struct {
	m           sync.Mutex
	cfgFilePath string
	username    string
}

func (c *config) SetUsername(username string) error {
	c.m.Lock()
	defer c.m.Unlock()

	file, err := os.OpenFile(c.cfgFilePath, os.O_WRONLY, readWriteFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, username)
	if err != nil {
		return err
	}

	c.username = username
	return nil
}

func (c *config) GetUsername() (string, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.username != "" {
		return c.username, nil
	}

	_, err := readFileContent(c.cfgFilePath, func(data string) error {
		c.username = data
		return nil
	})
	if err != nil {
		return "", err
	}

	return c.username, nil
}
