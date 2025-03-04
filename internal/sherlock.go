package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/KonstantinGasser/sherlock/security"
)

const (
	// querySplitPoint refers to the command line argument coming from the user
	// in the form of group@account and the separator used for it
	querySplitPoint = "@"
)

var (
	ErrNotSetup     = fmt.Errorf("sherlock needs to bee set-up first (use sherlock setup)")
	ErrNoSuchGroup  = fmt.Errorf("provided group cannot be found (use sherlock add group)")
	ErrWrongKey     = fmt.Errorf("wrong group key")
	ErrInvalidQuery = fmt.Errorf("invalid query. Query should be %q", "group@account")
)

type StateOption func(g *Group, acc string) error

func OptAddAccount(account *Account) StateOption {
	return func(g *Group, acc string) error {
		return g.append(account)
	}
}

// OptAccPassword returns a StateOption to change an account password
func OptAccPassword(password string, insecure bool) StateOption {
	return func(g *Group, acc string) error {
		account, err := g.lookup(acc)
		if err != nil {
			return err
		}
		if err := account.update(updateFieldPassword(password, insecure)); err != nil {
			return err
		}
		return nil
	}
}

// OptAccName returns a StateOption to change an account name
func OptAccName(name string) StateOption {
	return func(g *Group, acc string) error {
		if ok := g.exists(name); ok {
			return ErrAccountExists
		}
		account, err := g.lookup(acc)
		if err != nil {
			return err
		}
		if err := account.update(updateFieldName(name)); err != nil {
			return err
		}
		return nil
	}
}

func OptsAccTag(tag string) StateOption {
	return func(g *Group, acc string) error {
		account, err := g.lookup(acc)
		if err != nil {
			return err
		}
		if err := account.update(updateFieldTag(tag)); err != nil {
			return err
		}
		return nil
	}
}

// OptAccDelete returns a StateOption deleting an account if it exists
func OptAccDelete() StateOption {
	return func(g *Group, acc string) error {
		return g.delete(acc)
	}
}

// FileSystem declares the functions sherlock requires to
// interact with the underlying file system
type FileSystem interface {
	InitFs(initVault []byte) error
	CreateGroup(name string, initVault []byte) error
	GroupExists(name string) error
	VaultExists(group string) error
	ReadGroupVault(group string) ([]byte, error)
	Delete(ctx context.Context, gid string) error
	Write(ctx context.Context, gid string, data []byte) error
	ReadRegisteredGroups() ([]string, error)
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
func (sh *Sherlock) Setup(groupKey string) error {
	vault, err := security.InitWithDefault(groupKey, Group{
		GID:      "default",
		Accounts: make([]*Account, 0),
	})
	if err != nil {
		return err
	}

	if err := sh.fileSystem.InitFs(vault); err != nil {
		return err
	}
	return nil
}

// DeleteGroup irreversible deletes a group from sherlock
func (sh *Sherlock) DeleteGroup(ctx context.Context, gid string) error {
	return sh.fileSystem.Delete(ctx, gid)
}

// SetupGroup creates the group in the file system
// if the group does not already exists
func (sh Sherlock) SetupGroup(name string, groupKey string, insecure bool) error {
	if err := sh.GroupExists(name); err != nil {
		return err
	}
	group, err := NewGroup(name)
	if err != nil {
		return err
	}
	if !insecure {
		// check password strength for group key
		if err := group.secure(groupKey); err != nil {
			return err
		}
	}
	vault, err := security.InitWithDefault(groupKey, group)
	if err != nil {
		return err
	}
	return sh.fileSystem.CreateGroup(name, vault)
}

func (sh Sherlock) GroupExists(name string) error {
	return sh.fileSystem.GroupExists(name)
}

// ValidateGroupKey function validates the group's key for the requested groupID
func (sh *Sherlock) CheckGroupKey(ctx context.Context, query, groupKey string) error {
	gid, _, err := SplitQuery(query)
	if err != nil {
		return err
	}
	bytes, err := sh.fileSystem.ReadGroupVault(gid)
	if err != nil {
		return err
	}
	var group Group
	if err := security.DecryptVault(bytes, groupKey, &group); err != nil {
		return ErrWrongKey
	}
	return nil
}

// GetAccount looks up the requested account
// to locate an account the query needs to include the group
// like so group@account
func (sh Sherlock) GetAccount(query string, groupKey string) (*Account, error) {
	gid, name, err := SplitQuery(query)
	if err != nil {
		return nil, err
	}

	group, err := sh.LoadGroup(gid, groupKey)
	if err != nil {
		return nil, err
	}
	return group.lookup(name)
}

// UpdateState executes the passed in StateOption to perform state changes on a group
func (sh Sherlock) UpdateState(ctx context.Context, query, groupKey string, opt StateOption) error {
	gid, name, err := SplitQuery(query)
	if err != nil {
		return err
	}

	group, err := sh.LoadGroup(gid, groupKey)
	if err != nil {
		return err
	}
	if err := opt(group, name); err != nil {
		return err
	}
	return sh.WriteGroup(ctx, gid, groupKey, group)
}

// LoadGroup loads and decrypts the group vault
func (sh Sherlock) LoadGroup(gid string, groupKey string) (*Group, error) {
	bytes, err := sh.fileSystem.ReadGroupVault(gid)
	if err != nil {
		return nil, err
	}
	var group Group
	if err := security.DecryptVault(bytes, groupKey, &group); err != nil {
		return nil, ErrWrongKey
	}
	return &group, nil
}

// WriteGroup encrypts and write the group vault
func (sh Sherlock) WriteGroup(ctx context.Context, gid string, groupKey string, group *Group) error {
	serialized, err := group.serizalize()
	if err != nil {
		return err
	}
	encrypted, err := security.EncryptVault(serialized, groupKey)
	if err != nil {
		return err
	}
	return sh.fileSystem.Write(ctx, gid, encrypted)
}

// SplitQuery verifies that a query (for get,update command) are in the correct
// format: group@account
func SplitQuery(query string) (string, string, error) {
	set := strings.Split(query, querySplitPoint)
	if len(set) != 2 {
		return "", "", ErrInvalidQuery
	}
	return set[0], set[1], nil
}

// ReadRegisteredGroups loads saved groups
func (sh Sherlock) ReadRegisteredGroups() ([]string, error) {
	groups, err := sh.fileSystem.ReadRegisteredGroups()
	if err != nil {
		return nil, err
	}
	return groups, nil
}
