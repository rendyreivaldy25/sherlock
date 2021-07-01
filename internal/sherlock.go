package internal

import (
	"context"
	"fmt"

	"github.com/KonstantinGasser/sherlock/internal/security"
)

var (
	ErrNotSetup    = fmt.Errorf("sherlock needs to bee set-up first (use sherlock setup)")
	ErrNoSuchGroup = fmt.Errorf("provided group cannot be found (use sherlock add --group)")
)

// FileSystem declares the functions sherlock requires to
// interact with the underlying file system
type FileSystem interface {
	InitFs(initVault []byte) error
	CreateGroup(name string, initVault []byte) error
	GroupExists(name string) error
	VaultExists(group string) error
	ReadGroupVault(group string) ([]byte, error)
	WriteAccount(account *Account) error
}

type Sherlock struct {
	fileSystem FileSystem
}

// New return new Sherlock instance
func NewSherlock(fs FileSystem) *Sherlock {
	return &Sherlock{
		fileSystem: fs,
	}
}

func (sh Sherlock) IsSetUp() error {
	if err := sh.fileSystem.GroupExists("default"); err == nil { // default group does not exists
		return ErrNotSetup
	}
	if err := sh.fileSystem.VaultExists("default"); err == nil {
		return ErrNotSetup
	}
	return nil
}

// Setup checks if a main password for the vault has already been
// set which is required for every further command. Setup will create required directories
// if those are missing
func (sh *Sherlock) Setup(partionKey string) error {
	vault, err := security.InitWithDefault(partionKey)
	if err != nil {
		return err
	}

	if err := sh.fileSystem.InitFs(vault); err != nil {
		return err
	}
	return nil
}

// SetupGroup creates the group partition in the file system
// if the group does not already exists
func (sh Sherlock) SetupGroup(name string, partionKey string) error {
	if err := sh.GroupExists(name); err != nil {
		return err
	}

	vault, err := security.InitWithDefault(partionKey)
	if err != nil {
		return err
	}
	return sh.fileSystem.CreateGroup(name, vault)
}

func (sh Sherlock) GroupExists(name string) error {
	return sh.fileSystem.GroupExists(name)
}

func (sh *Sherlock) AddAccount(ctx context.Context, account *Account) error {
	return fmt.Errorf("sherlock.AddAccount: not implemented")
}
