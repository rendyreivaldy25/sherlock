package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	prettyDateLayout = "Monday, 02. January 2006"
)

var (
	ErrAccountExists = fmt.Errorf("account for group already exists")
	ErrNoSuchAccount = fmt.Errorf("account not found")
)

// Group groups Accounts
type Group struct {
	GID      string     `json:"name"`
	Accounts []*Account `json:"accounts"`
}

func NewGroup(name string) *Group {
	return &Group{
		GID:      name,
		Accounts: make([]*Account, 0),
	}
}

// append appends an account to a group if it does not already exists
func (g *Group) append(account *Account) error {
	if ok := g.exists(account); ok {
		return ErrAccountExists
	}
	g.Accounts = append(g.Accounts, account)
	return nil
}

func (g Group) lookup(accountName string) (*Account, error) {
	for _, a := range g.Accounts {
		if a.Name == accountName {
			return a, nil
		}
	}
	return nil, ErrNoSuchAccount
}

// delete deletes a given account from the group, returns an ErrNoSuchAccount
// if account not present
func (g *Group) delete(account string) error {
	var offset *int
	for i, a := range g.Accounts {
		if a.Name == account {
			offset = &i
		}
	}
	if offset == nil {
		return ErrNoSuchAccount
	}

	g.Accounts = append(g.Accounts[:*offset], g.Accounts[*offset+1:]...)
	return nil
}

// exists checks an account is already present in the group
// the Account.Name and Account.Tag field build the pk for an Account
func (g Group) exists(account *Account) bool {
	for _, a := range g.Accounts {
		if account.Name == a.Name && account.Tag == a.Tag {
			return true
		}
	}
	return false
}

func (g Group) serizalize() ([]byte, error) {
	return json.Marshal(g)
}

// Table builds the Group in such a way that it can be consumed by the tablewriter.Table
func (g Group) Table() [][]string {
	var accounts = make([][]string, len(g.Accounts))
	for i, item := range g.Accounts {
		accounts[i] = []string{
			g.GID,
			item.Name,
			strings.Join([]string{"#", item.Tag}, ""),
			item.CreatedOn.Format(prettyDateLayout),
		}
	}
	return accounts
}
