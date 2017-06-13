package config

import (
	"errors"
	"strconv"
	"strings"
)

// parseKVSlice parses a configuration string in the form
//
//   key=val;key=val,key=val;key=val
//
// into a list of string maps. maps are separated by comma and key/value
// pairs within a map are separated by semicolons. The first key/value
// pair of a map can omit the key and its value will be stored under the
// empty key. This allows support of legacy configuration formats which
// are
//
//   val;opt1=val1;opt2=val2;...
func parseKVSlice(in string) ([]map[string]string, error) {
	var keyOrFirstVal string
	maps := []map[string]string{}
	m := map[string]string{}

	newMap := func() {
		if len(m) > 0 {
			maps = append(maps, m)
			m = map[string]string{}
		}
	}

	v := ""
	s := []rune(in)
	state := stateFirstKey
	for {
		if len(s) == 0 {
			break
		}
		typ, val, n := lex(s)
		s = s[n:]
		// fmt.Println("parse:", "typ:", typ, "val:", val, "v:", v, "state:", string(state), "s:", string(s))
		switch state {
		case stateFirstKey:
			switch typ {
			case itemText:
				keyOrFirstVal = strings.TrimSpace(val)
				state = stateAfterFirstKey
			case itemComma, itemSemicolon:
				continue
			default:
				return nil, errors.New(val)
			}

		// the first value is allowed to omit the key
		// a=b;c=d and b;c=d are valid
		case stateAfterFirstKey:
			switch typ {
			case itemEqual:
				state = stateVal
			case itemComma:
				if keyOrFirstVal != "" {
					m[""] = keyOrFirstVal
				}
				newMap()
				state = stateFirstKey
			case itemSemicolon:
				if keyOrFirstVal != "" {
					m[""] = keyOrFirstVal
				}
				state = stateKey
			default:
				return nil, errors.New(val)
			}

		case stateKey:
			switch typ {
			case itemText:
				keyOrFirstVal = strings.TrimSpace(val)
				state = stateEqual
			case itemComma, itemSemicolon:
				continue
			default:
				return nil, errors.New(val)
			}

		case stateEqual:
			switch typ {
			case itemEqual:
				state = stateVal
			default:
				return nil, errors.New(val)
			}

		case stateVal:
			switch typ {
			case itemText, itemEqual:
				v += val
			case itemComma:
				m[keyOrFirstVal] = v
				v = ""
				newMap()
				state = stateFirstKey
			case itemSemicolon:
				m[keyOrFirstVal] = v
				v = ""
				state = stateKey
			default:
				return nil, errors.New(val)
			}
		}
	}
	switch state {
	case stateVal:
		m[keyOrFirstVal] = v
	case stateAfterFirstKey:
		if keyOrFirstVal != "" {
			m[""] = keyOrFirstVal
		}
	}
	if len(m) > 0 {
		maps = append(maps, m)
	}
	if len(maps) == 0 {
		return nil, nil
	}
	return maps, nil
}

type itemType string

const (
	itemText      itemType = "TEXT"
	itemEqual              = "EQUAL"
	itemSemicolon          = "SEMICOLON"
	itemComma              = "COMMA"
	itemError              = "ERROR"
)

func (t itemType) String() string {
	return string(t)
}

type state string

const (

	// lexer states
	stateStart    state = "start"
	stateText           = "text"
	stateQText          = "qtext"
	stateQTextEnd       = "qtextend"
	stateQTextEsc       = "qtextesc"

	// parser states
	stateFirstKey      = "first-key"
	stateKey           = "key"
	stateEqual         = "equal"
	stateVal           = "val"
	stateAfterFirstKey = "equal-comma-semicolon"
)

func lex(s []rune) (itemType, string, int) {
	isComma := func(r rune) bool { return r == ',' }
	isSemicolon := func(r rune) bool { return r == ';' }
	isEqual := func(r rune) bool { return r == '=' }
	isEscape := func(r rune) bool { return r == '\\' }
	isQuote := func(r rune) bool { return r == '"' || r == '\'' }

	var quote rune
	state := stateStart
	for i, r := range s {
		// fmt.Println("lex:", "i:", i, "r:", string(r), "state:", string(state))
		switch state {
		case stateStart:
			switch {
			case isComma(r):
				return itemComma, string(r), 1
			case isSemicolon(r):
				return itemSemicolon, string(r), 1
			case isEqual(r):
				return itemEqual, string(r), 1
			case isQuote(r):
				quote = r
				state = stateQText
			default:
				state = stateText
			}
		case stateText:
			switch {
			case isComma(r) || isSemicolon(r) || isEqual(r):
				return itemText, string(s[:i]), i
			default:
				// state = stateText
			}
		case stateQText:
			switch {
			case r == quote:
				state = stateQTextEnd
			case isEscape(r):
				state = stateQTextEsc
			default:
				// state = stateQText
			}

		case stateQTextEsc:
			state = stateQText

		case stateQTextEnd:
			v, err := strconv.Unquote(string(s[:i]))
			if err != nil {
				return itemError, "invalid escape sequence", i
			}
			return itemText, v, i
		}
	}

	// fmt.Println("lex:", "state:", string(state))
	switch state {
	case stateQText:
		return itemError, "unbalanced quotes", len(s)
	case stateQTextEsc:
		return itemError, "unterminated escape sequence", len(s)
	case stateQTextEnd:
		v, err := strconv.Unquote(string(s))
		if err != nil {
			return itemError, "invalid escape sequence", len(s)
		}
		return itemText, v, len(s)
	default:
		return itemText, string(s), len(s)
	}
}
