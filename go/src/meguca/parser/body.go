// Package parser parses and verifies user-sent post data
package parser

import (
	"fmt"
	"meguca/common"
	"meguca/util"
	"regexp"
	"unicode"
)

var (
	linkRegexp = regexp.MustCompile(`^>{2,}(\d+)$`)
)

// Rune is non-printable
type ErrNonPrintable rune

func (e ErrNonPrintable) Error() string {
	return fmt.Sprintf("contains non-printable character: %d", int(e))
}

// Needed to avoid cyclic imports for the 'db' package
func init() {
	common.ParseBody = ParseBody
}

// ParseBody parses the entire post text body for commands and links.
// internal: function was called by automated upkeep task
func ParseBody(body []byte, board string, internal bool) (
	links []common.Link, com []common.Command, err error,
) {
	err = IsPrintableString(string(body), true)
	if err != nil {
		if internal {
			err = nil
			// Strip any non-printables for automated post closing
			s := make([]byte, 0, len(body))
			for _, r := range []rune(string(body)) {
				if IsPrintable(r, true) == nil {
					s = append(s, string(r)...)
				}
			}
			body = s
		} else {
			return
		}
	}

	start := 0
	lineStart := 0

	for i, b := range body {
		switch b {
		case '\n', ' ', '\t':
		default:
			if i == len(body)-1 {
				i++
			} else {
				continue
			}
		}

		_, word, _ := util.SplitPunctuation(body[start:i])
		start = i + 1
		if len(word) == 0 {
			goto next
		}

		switch word[0] {
		case '>':
			m := linkRegexp.FindSubmatch(word)
			if m == nil {
				goto next
			}
			var l common.Link
			l, err = parseLink(m)
			switch {
			case err != nil:
				return
			case l.ID != 0:
				links = append(links, l)
			}
		case '#':
			if body[lineStart] == '>' { // Ignore hash commands in quotes
				goto next
			}
			m := common.CommandRegexp.FindSubmatch(word)
			if m == nil {
				goto next
			}
			var c common.Command
			c, err = parseCommand(m[1], board)
			switch err {
			case nil:
				com = append(com, c)
			case errTooManyRolls, errDieTooBig:
				// Consider command invalid
				err = nil
			default:
				return
			}
		}
	next:
		if b == '\n' {
			lineStart = i + 1
		}
	}

	return
}

// Checks, if r is printable.
// Also accepts tabs, and newlines, if multiline = true.
func IsPrintable(r rune, multiline bool) error {
	switch r {
	case '\t', '\n', 12288: // Japanese space
		if !multiline {
			return ErrNonPrintable(r)
		}
	default:
		if !unicode.IsPrint(r) {
			return ErrNonPrintable(r)
		}
	}
	return nil
}

// Checks, if all of s is printable.
// Also accepts tabs, and newlines, if multiline = true.
func IsPrintableString(s string, multiline bool) error {
	for _, r := range []rune(s) {
		if err := IsPrintable(r, multiline); err != nil {
			return err
		}
	}
	return nil
}

// Checks, if all of s is printable.
// Also accepts tabs, and newlines, if multiline = true.
func IsPrintableRunes(s []rune, multiline bool) error {
	for _, r := range s {
		if err := IsPrintable(r, multiline); err != nil {
			return err
		}
	}
	return nil
}
