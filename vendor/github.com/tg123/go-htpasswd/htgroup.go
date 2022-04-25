// Package htpasswd groups provides an autorisation mechanism using Apache-style group files.
//
// An Apache group file looks like this:
// users: user1 user2 user3
// admins: user1
//
// Basic usage of this package:
//
// userGroups, groupLoadErr := htgroup.NewGroups("./my-group-file", nil)
// ok := userGroups.IsUserInGroup(username, "admins")
package htpasswd


import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)


// Data structure for users and theirs groups (map).
// The map key is the user, the value is an array of groups.
type userGroupMap map[string][]string


// A HTGroup encompasses an Apache-style group file.
type HTGroup struct {
	filePath string
	mutex sync.Mutex
	userGroups userGroupMap
}


// NewGroups creates a HTGroup from an Apache-style group file.
//
// The filename must exist and be accessible to the process, as well as being a valid group file.
//
// bad is a function, which if not nil will be called for each malformed or rejected entry in the group file.
func NewGroups(filename string, bad BadLineHandler) (*HTGroup, error) {
	htGroup := HTGroup {
		filePath: filename,
	}
	return &htGroup, htGroup.ReloadGroups(bad)
}


// NewGroupsFromReader is like NewGroups but reads from r instead of a named file.
func NewGroupsFromReader(r io.Reader, bad BadLineHandler) (*HTGroup, error) {
	htGroup := HTGroup {}

	readFileErr := htGroup.ReloadGroupsFromReader(r, bad)
	if readFileErr != nil {
		return nil, readFileErr
	}

	return &htGroup, nil
}


// ReloadGroups rereads the group file.
func (htGroup *HTGroup) ReloadGroups(bad BadLineHandler) error {
	htGroup.mutex.Lock()
	filename := htGroup.filePath
	htGroup.mutex.Unlock()
	file, err := os.Open(filename)
    if err != nil {
		return err
    }
	defer file.Close()

	return htGroup.ReloadGroupsFromReader(file, bad)
}


// ReloadGroupsFromReader rereads the group file from a Reader.
func (htGroup *HTGroup) ReloadGroupsFromReader(r io.Reader, bad BadLineHandler) error {
	userGroups := make(userGroupMap)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if lineErr := processLine(&userGroups, line); lineErr != nil && bad != nil {
			bad(lineErr)
		}
    }
	if scannerErr := scanner.Err(); scannerErr != nil {
		return fmt.Errorf("Error scanning group file: %s", scannerErr.Error())
	}

	htGroup.mutex.Lock()
	htGroup.userGroups = userGroups
	htGroup.mutex.Unlock()

	return nil
}


func processLine(userGroups *userGroupMap, rawLine string) error {
	line := strings.TrimSpace(rawLine)
	if line == "" {
		return nil
	}

	groupAndUsers := strings.SplitN(line, ":", 2)
	if len(groupAndUsers) != 2 {
		return fmt.Errorf("malformed line, no colon: %s", line)
	}

	var group = strings.TrimSpace(groupAndUsers[0]);
	var users = strings.Fields(groupAndUsers[1])
	for _, user := range users {
		if (*userGroups)[user] == nil {
			(*userGroups)[user] = []string {}
		}
		(*userGroups)[user] = append((*userGroups)[user], group)
	}

	return nil
}


// IsUserInGroup checks whether the user is in a group.
// Returns true of user is in that group, otherwise false.
func (htGroup *HTGroup) IsUserInGroup(user string, group string) bool {
	groups := htGroup.GetUserGroups(user)
	return containsGroup(groups, group)
}


// GetUserGroups reads all groups of a user.
// Returns all groups as a string array or an empty array.
func (htGroup *HTGroup) GetUserGroups(user string) []string {
	htGroup.mutex.Lock()
	groups := htGroup.userGroups[user]
	htGroup.mutex.Unlock()

	if (groups == nil) {
		return []string {}
	}
	return groups
}


func containsGroup(groups []string, group string) bool {
    for _, g := range groups {
        if g == group {
            return true
        }
    }
    return false
}
