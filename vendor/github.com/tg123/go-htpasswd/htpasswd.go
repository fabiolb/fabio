// Package htpasswd provides HTTP Basic Authentication using Apache-style htpasswd files
// for the user and password data.
//
// It supports most common hashing systems used over the decades and can be easily extended
// by the programmer to support others. (See the sha.go source file as a guide.)
//
// You will want to use something like...
//      myauth := htpasswd.New("My Realm", "./my-htpasswd-file", htpasswd.DefaultSystems, nil)
//      m.Use(myauth.Handler)
// ...to configure your authentication and then use the myauth.Handler as a middleware handler in your Martini stack.
// You should read about that nil, as well as Reread() too.
package htpasswd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// An EncodedPasswd is created from the encoded password in a password file by a PasswdParser.
//
// The password files consist of lines like "user:passwd-encoding". The user part is stripped off and
// the passwd-encoding part is captured in an EncodedPasswd.
type EncodedPasswd interface {
	// Return true if the string matches the password.
	// This may cache the result in the case of expensive comparison functions.
	MatchesPassword(pw string) bool
}

// PasswdParser examines an encoded password, and if it is formatted correctly and sane, return an
// EncodedPasswd which will recognize it.
//
// If the format is not understood, then return nil
// so that another parser may have a chance. If the format is understood but not sane,
// return an error to prevent other formats from possibly claiming it
//
// You may write and supply one of these functions to support a format (e.g. bcrypt) not
// already included in this package. Use sha.c as a template, it is simple but not too simple.
type PasswdParser func(pw string) (EncodedPasswd, error)

type passwdTable map[string]EncodedPasswd

// A BadLineHandler is used to notice bad lines in a password file. If not nil, it will be
// called for each bad line with a descriptive error. Think about what you do with these, they
// will sometimes contain hashed passwords.
type BadLineHandler func(err error)

// An HtpasswdFile encompasses an Apache-style htpasswd file for HTTP Basic authentication
type HtpasswdFile struct {
	filePath string
	mutex    sync.Mutex
	passwds  passwdTable
	parsers  []PasswdParser
}

// DefaultSystems is an array of PasswdParser including all builtin parsers. Notice that Plain is last, since it accepts anything
var DefaultSystems = []PasswdParser{AcceptMd5, AcceptSha, AcceptBcrypt, AcceptSsha, AcceptPlain}

// New creates an HtpasswdFile from an Apache-style htpasswd file for HTTP Basic Authentication.
//
// The realm is presented to the user in the login dialog.
//
// The filename must exist and be accessible to the process, as well as being a valid htpasswd file.
//
// parsers is a list of functions to handle various hashing systems. In practice you will probably
// just pass htpasswd.DefaultSystems, but you could make your own to explicitly reject some formats or
// implement your own.
//
// bad is a function, which if not nil will be called for each malformed or rejected entry in
// the password file.
func New(filename string, parsers []PasswdParser, bad BadLineHandler) (*HtpasswdFile, error) {
	bf := HtpasswdFile{
		filePath: filename,
		parsers:  parsers,
	}

	if err := bf.Reload(bad); err != nil {
		return nil, err
	}

	return &bf, nil
}

// Match checks the username and password combination to see if it represents
// a valid account from the htpassword file.
func (bf *HtpasswdFile) Match(username, password string) bool {
	bf.mutex.Lock()
	matcher, ok := bf.passwds[username]
	bf.mutex.Unlock()

	if ok && matcher.MatchesPassword(password) {
		// we are good
		return true
	}

	return false
}

// Reload rereads the htpassword file..
// You will need to call this to notice any changes to the password file.
// This function is thread safe. Someone versed in fsnotify might make it
// happen automatically. Likewise you might also connect a SIGHUP handler to
// this function.
func (bf *HtpasswdFile) Reload(bad BadLineHandler) error {
	// with the file...
	f, err := os.Open(bf.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// ... and a new map ...
	newPasswdMap := passwdTable{}

	// ... for each line ...
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// ... add it to the map, noting errors along the way
		if perr := bf.addHtpasswdUser(&newPasswdMap, line); perr != nil && bad != nil {
			bad(perr)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error scanning htpasswd file: %s", err.Error())
	}

	// .. finally, safely swap in the new map
	bf.mutex.Lock()
	bf.passwds = newPasswdMap
	bf.mutex.Unlock()

	return nil
}

// addHtpasswdUser processes a line from an htpasswd file and add it to the user/password map. We may
// encounter some malformed lines, this will not be an error, but we will log them if
// the caller has given us a logger.
func (bf *HtpasswdFile) addHtpasswdUser(pwmap *passwdTable, rawLine string) error {
	// ignore white space lines
	line := strings.TrimSpace(rawLine)
	if line == "" {
		return nil
	}

	// split "user:encoding" at colon
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("malformed line, no colon: %s", line)
	}

	user := parts[0]
	encoding := parts[1]

	// give each parser a shot. The first one to produce a matcher wins.
	// If one produces an error then stop (to prevent Plain from catching it)
	for _, p := range bf.parsers {
		matcher, err := p(encoding)
		if err != nil {
			return err
		}
		if matcher != nil {
			(*pwmap)[user] = matcher
			return nil // we are done, we took to first match
		}
	}

	// No one liked this line
	return fmt.Errorf("unable to recognize password for %s in %s", user, encoding)
}
