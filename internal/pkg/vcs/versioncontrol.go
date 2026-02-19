package vcs

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const (
	configFile = "config.txt"
)

var (
	ErrConfigFileNotCreated = errors.New("config file not created")
)

type VersionControl struct {
	m   sync.Mutex
	cfg *config

	vcsRoot      string
	currentRepo  *Repository
	repositories map[string]*Repository
}

func NewVersionControl(dir string) (*VersionControl, error) {
	if err := createDirectory(dir); err != nil {
		return nil, err
	}

	if err := createFile(filepath.Join(dir, configFile)); err != nil {
		return nil, ErrConfigFileNotCreated
	}

	vc := &VersionControl{
		cfg:          &config{cfgFilePath: filepath.Join(dir, configFile)},
		vcsRoot:      dir,
		repositories: make(map[string]*Repository),
	}

	if err := vc.initRepositories(); err != nil {
		return nil, err
	}

	return vc, nil
}

func (vc *VersionControl) GetWorkingRepository() *Repository {
	return vc.currentRepo
}

func (vc *VersionControl) ListRepositories() []string {
	var result []string
	for name := range vc.repositories {
		result = append(result, name)
	}
	return result
}

func (vc *VersionControl) ReadConfig() (string, error) {
	return vc.cfg.GetUsername()
}

func (vc *VersionControl) WriteConfig(username string) error {
	return vc.cfg.SetUsername(username)
}

func (vc *VersionControl) Checkout(dir string) error {
	if repo, ok := vc.repositories[dir]; ok {
		vc.currentRepo = repo
		return nil
	}

	repository, err := newRepository(vc.vcsRoot, dir)
	if err != nil {
		return err
	}

	vc.m.Lock()
	defer vc.m.Unlock()

	vc.repositories[dir] = repository
	vc.currentRepo = repository
	return nil
}

func (vc *VersionControl) initRepositories() error {
	folders, err := listFolders(vc.vcsRoot)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		repository, err := newRepository(vc.vcsRoot, folder)
		if err != nil {
			return err
		}

		vc.repositories[folder] = repository
	}

	return nil
}

func listFolders(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var folders []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != vcsDir {
			folders = append(folders, e.Name())
		}
	}

	return folders, nil
}
