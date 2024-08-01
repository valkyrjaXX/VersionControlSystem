package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	vcsDirPath = "./vcs"
)

var (
	rootDir *RootDir
)

func init() {
	dir, err := NewRootDir(vcsDirPath)
	if err != nil {
		panic(err)
	}

	rootDir = dir
}

func main() {
	var help bool
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()

	//fmt.Println(os.Args)

	if len(os.Args) == 1 || help {
		fmt.Println(Help.Description)
		return
	}

	var commandName string
	if len(os.Args) > 1 {
		commandName = os.Args[1]
	}

	command, err := CommandOfName(commandName)
	if err != nil {
		fmt.Println(fmt.Sprintf("'%s' is not a SVCS command.", commandName))
		return
	}

	if command.Execute == nil {
		fmt.Println(command.Description)
		return
	}

	args := os.Args[2:]
	if err := command.Execute(args...); err != nil {
		fmt.Println(err)
	}
}

// commands

var (
	ErrCommandNotSupported = errors.New("command not supported")
)

type CommandDef struct {
	Name        string
	Description string
	Execute     Executor
}

type Executor func(args ...string) error

var Commands = []*CommandDef{Config, Add, Log, Commit, Checkout}

func CommandOfName(name string) (*CommandDef, error) {
	if len(name) == 0 || name == Help.Name {
		return Help, nil
	}

	for _, cmdDef := range Commands {
		if cmdDef.Name == name {
			return cmdDef, nil
		}
	}

	return nil, ErrCommandNotSupported
}

// help

var Help = &CommandDef{
	Name:        "help",
	Description: buildHelpDescription(),
}

func buildHelpDescription() string {
	text := []string{"These are SVCS commands:"}

	for _, cmdDef := range Commands {
		text = append(text, fmt.Sprintf("%s %s", cmdDef.Name, cmdDef.Description))
	}

	return strings.Join(text, "\n")
}

// config

var Config = &CommandDef{
	Name:        "config",
	Description: "Get and set a username.",
	Execute: func(args ...string) error {
		if len(args) == 0 {
			config, err := rootDir.ReadConfig()
			if err != nil {
				return err
			}

			if len(config) == 0 {
				fmt.Println("Please, tell me who you are.")
				return nil
			}

			fmt.Println(fmt.Sprintf("The username is %s.", strings.Trim(config, "\n")))
			return nil
		}

		username := args[0]
		if err := rootDir.WriteConfig(username); err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("The username is %s.", username))
		return nil
	},
}

// add

var Add = &CommandDef{
	Name:        "add",
	Description: "Add a file to the index.",
	Execute: func(args ...string) error {
		if len(args) == 0 {
			var titlePrinted bool
			err := rootDir.ReadIndex(func(data string) error {
				if !titlePrinted {
					fmt.Println("Tracked files:")
					titlePrinted = true
				}

				fmt.Println(data)
				return nil
			})
			if err != nil {
				if errors.Is(err, ErrEmptyFile) {
					fmt.Println("Add a file to the index.")
					return nil
				}

				return err
			}

			return nil
		}

		filename := args[0]
		fileExist, err := fileExists(filename)
		if err != nil {
			return err
		}

		if !fileExist {
			fmt.Println(fmt.Sprintf("Can't find '%s'.", filename))
			return nil
		}

		if err := rootDir.WriteToIndex(filename); err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("The file '%s' is tracked.", filename))
		return nil
	},
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// log

var Log = &CommandDef{
	Name:        "log",
	Description: "Show commit logs.",
	Execute: func(args ...string) error {
		err := rootDir.ReadLog(func(commit, author, comment string) {
			defer func() {
				fmt.Println(fmt.Sprintf("commit %s", commit))
				fmt.Println(fmt.Sprintf("Author: %s", author))
				fmt.Println(comment)
			}()
		})
		if err != nil {
			if errors.Is(err, ErrEmptyFile) {
				fmt.Println("No commits yet.")
				return nil
			}

			return err
		}

		return nil
	},
}

// commit

var Commit = &CommandDef{
	Name:        "commit",
	Description: "Save changes.",
	Execute: func(args ...string) error {
		if len(args) == 0 {
			fmt.Println("Message was not passed.")
			return nil
		}

		message := args[0]
		commit, err := rootDir.CreateCommitPromise(message)
		if err != nil {
			return err
		}

		err = rootDir.ReadIndex(func(data string) error {
			return commit.CopyFile(data)
		})
		if err != nil {
			if errors.Is(err, ErrEmptyFile) {
				fmt.Println("Nothing to commit.")
				return nil
			}

			return err
		}

		commited, err := rootDir.Commit(commit)
		if err != nil {
			return err
		}

		if !commited {
			fmt.Println("Nothing to commit.")
			return nil
		}

		fmt.Println("Changes are committed.")
		return nil
	},
}

// checkout

var Checkout = &CommandDef{
	Name:        "checkout",
	Description: "Restore a file.",
	Execute: func(args ...string) error {
		if len(args) == 0 {
			fmt.Println("Commit id was not passed.")
			return nil
		}

		commit := args[0]
		err := rootDir.Checkout(commit)
		if err != nil {
			if errors.Is(err, ErrCommitNotFound) {
				fmt.Println("Commit does not exist.")
				return nil
			}
			return err
		}

		fmt.Println(fmt.Sprintf("Switched to commit %s.", commit))
		return nil
	},
}

// directory

const (
	commitsDir = "commits"

	configFile = "config.txt"
	indexFile  = "index.txt"
	logFile    = "log.txt"

	comma = ","
)

var (
	ErrEmptyFile      = errors.New("empty file")
	ErrCommitNotFound = errors.New("commit not found")

	readWriteFileMode = os.FileMode(0444)
)

type CommitPromise struct {
	Uuid        UUID
	Hash        hash.Hash
	TempDirPath string
	Username    string
	Message     string
	dirCreated  bool
}

func (c *CommitPromise) CopyFile(srcFilePath string) error {
	if !c.dirCreated {
		err := os.MkdirAll(c.TempDirPath, readWriteFileMode)
		if err != nil {
			return err
		}
	}

	filename := filepath.Base(srcFilePath)
	dstFilePath := filepath.Join(c.TempDirPath, filename)

	r, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := io.Copy(c.Hash, r); err != nil {
		return err
	}

	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	w, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}

	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

type RootDir struct {
	vcsDirPath string

	username string
}

func (r *RootDir) ReadConfig() (string, error) {
	if r.username != "" {
		return r.username, nil
	}

	_, err := r.readFileContent(configFile, func(data string) error {
		r.username = data
		return nil
	}, false)
	if err != nil {
		return "", err
	}

	return r.username, nil
}

func (r *RootDir) WriteConfig(username string) error {
	filePath := filepath.Join(r.vcsDirPath, configFile)
	file, err := os.OpenFile(filePath, os.O_WRONLY, readWriteFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, username)
	if err != nil {
		return err
	}

	r.username = username
	return nil
}

func (r *RootDir) ReadIndex(callback func(data string) error) error {
	read, err := r.readFileContent(indexFile, callback, false)
	if err != nil {
		return err
	}

	if read == 0 {
		return ErrEmptyFile
	}

	return nil
}

func (r *RootDir) WriteToIndex(fileToAdd string) error {
	filePath := filepath.Join(r.vcsDirPath, indexFile)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, readWriteFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, fileToAdd)
	return err
}

func (r *RootDir) ReadLog(callback func(commit, author, comment string)) error {
	read, err := r.readFileContent(logFile, func(data string) error {
		dataSlice := strings.SplitN(data, comma, 3)
		callback(dataSlice[0], dataSlice[1], dataSlice[2])
		return nil
	}, true)
	if err != nil {
		return err
	}

	if read == 0 {
		return ErrEmptyFile
	}

	return nil
}

func (r *RootDir) CreateCommitPromise(message string) (*CommitPromise, error) {
	uuid, err := generateUuid()
	if err != nil {
		return nil, err
	}

	username, err := r.ReadConfig()
	if err != nil {
		return nil, err
	}

	return &CommitPromise{
		Uuid:        uuid,
		Hash:        sha256.New(),
		Username:    username,
		TempDirPath: filepath.Join(r.vcsDirPath, commitsDir, "temp-"+string(uuid)),
		Message:     message,
	}, nil
}

func (r *RootDir) Commit(c *CommitPromise) (bool, error) {
	changesHash := fmt.Sprintf("%x", c.Hash.Sum(nil))
	commitDirPath := filepath.Join(r.vcsDirPath, commitsDir, changesHash)
	_, err := os.Stat(commitDirPath)
	if err == nil {
		return false, os.RemoveAll(c.TempDirPath)
	}

	if errors.Is(err, os.ErrNotExist) {
		err := os.Rename(c.TempDirPath, commitDirPath)
		if err != nil {
			return false, err
		}
	}

	err = r.writeLog(changesHash, c.Username, c.Message)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RootDir) Checkout(commitID string) error {
	commitDirPath := filepath.Join(r.vcsDirPath, commitsDir, commitID)
	_, err := os.Stat(commitDirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrCommitNotFound
		}
		return err
	}

	err = r.ReadIndex(func(dstFilePath string) error {
		srcFilePath := filepath.Join(commitDirPath, filepath.Base(dstFilePath))
		_, err := os.Stat(srcFilePath)
		if err != nil {
			return err
		}

		reader, err := os.Open(srcFilePath)
		if err != nil {
			return err
		}
		defer reader.Close()

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

	return err
}

func (r *RootDir) writeLog(hash, username, message string) error {
	filePath := filepath.Join(r.vcsDirPath, logFile)
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

func (r *RootDir) readFileContent(filename string, callback func(data string) error, reversed bool) (int, error) {
	filePath := filepath.Join(r.vcsDirPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var readErr error
	var read int
	for scanner.Scan() {
		text := scanner.Text()
		if reversed {
			defer func() {
				readErr = callback(text)
			}()
		} else {
			readErr = callback(text)
		}

		read += len([]byte(text))
	}

	if readErr != nil {
		return read, readErr
	}

	return read, scanner.Err()
}

func NewRootDir(root string) (*RootDir, error) {
	_, err := os.Stat(root)
	if err == nil {
		return &RootDir{vcsDirPath: vcsDirPath}, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		err = createRepoStructure(root)
		if err != nil {
			return nil, err
		}

		return &RootDir{
			vcsDirPath: vcsDirPath,
		}, nil
	}

	return nil, err
}

func createRepoStructure(rootDir string) error {
	err := os.MkdirAll(filepath.Join(rootDir, commitsDir), readWriteFileMode)
	if err != nil {
		return err
	}

	err = createInfraFiles(rootDir, configFile, indexFile, logFile)
	if err != nil {
		return err
	}

	return nil
}

func createInfraFiles(rootDir string, filenames ...string) error {
	for _, filename := range filenames {
		filePath := filepath.Join(rootDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		file.Close()
	}

	return nil
}

// uuid

type UUID string

func generateUuid() (UUID, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return UUID(uuid), nil
}
