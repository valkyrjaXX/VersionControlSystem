package vcs

import (
	"errors"
	"path/filepath"
	"sync"
)

const (
	configFile = "config.txt"
)

var (
	ErrDirectoryNotEmpty    = errors.New("directory is not empty")
	ErrConfigFileNotCreated = errors.New("config file not created")
)

type VersionControl struct {
	m   sync.Mutex
	cfg *config

	vcsRoot      string
	currentRepo  string
	repositories map[string]*Repository
}

func NewVersionControl(dir string) (*VersionControl, error) {
	if err := createDirectory(dir); err != nil {
		return nil, err
	}

	if err := createFile(filepath.Join(dir, configFile)); err != nil {
		return nil, ErrConfigFileNotCreated
	}

	return &VersionControl{
		cfg:          &config{cfgFilePath: filepath.Join(dir, configFile)},
		vcsRoot:      dir,
		repositories: make(map[string]*Repository),
	}, nil
}

func (vc *VersionControl) GetWorkingRepository() string {
	return vc.currentRepo
}

func (vc *VersionControl) ReadConfig() (string, error) {
	return vc.cfg.GetUsername()
}

func (vc *VersionControl) WriteConfig(username string) error {
	return vc.cfg.SetUsername(username)
}

func (vc *VersionControl) Checkout(dir string) error {
	if _, ok := vc.repositories[dir]; ok {
		vc.currentRepo = dir
		return nil
	}

	repository, err := newRepository(vc.vcsRoot, dir)
	if err != nil {
		return err
	}

	vc.m.Lock()
	defer vc.m.Unlock()

	vc.repositories[dir] = repository
	vc.currentRepo = dir
	return nil
}
