package main

import (
	"bufio"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"io"
	"os"
	"strings"
)

var errMissingClosingBracket = fmt.Errorf("missing closing bracket")

func populateInventory(reader *bufio.Reader, i *inventory.Inventory) error {
	var curCmd string
	var curValue string

	prefix := make([]string, 0, 10)
	lineNo := 1

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			// If we reached the end of the file no error should be returned.
			if err == io.EOF {
				return nil
			}

			return fmt.Errorf("reading file at line %d: %w", lineNo, err)
		}

		switch r {
		case '{':
			// parse start section
			if curValue != "" {
				curValue = strings.Replace(curValue, "/", ":", -1)
				prefix = append(prefix, fmt.Sprintf("%s:%s", curCmd, strings.Trim(curValue, " \t")))
			} else {
				prefix = append(prefix, curCmd)
			}
			curCmd = ""
			curValue = ""
		case '}':
			// parse end section
			closeIdx := len(prefix) - 1
			if closeIdx < 0 {
				return fmt.Errorf("at line %d: %w", lineNo, errMissingClosingBracket)
			}
			prefix = prefix[:closeIdx]
		case ';':
			// parse end statement
			prefix = append(prefix, curCmd)
			err = i.SetItem(strings.Join(prefix, "/"), "value", curValue)
			if err != nil {
				return err
			}
			prefix = prefix[:len(prefix)-1]

			curValue = ""
			curCmd = ""
		case '\n':
			// parse end line and subsequent spaces
			for err == nil && (r == '\n' || r == ' ' || r == '\t') {
				if r == '\n' {
					lineNo++
				}
				r, _, err = reader.ReadRune()
			}
			if err != nil {
				continue // Break to outer loop so we can handle errors/EOF
			}
			err = reader.UnreadRune()
			if err != nil {
				return fmt.Errorf("parsing line %d: %w", lineNo, err)
			}
		case '#':
			// ignore comments
			for r != '\n' {
				r, _, _ = reader.ReadRune()
			}
		case '\t', ' ':
			if curValue == "" {
				continue
			}
			if curValue != "" && curCmd == "" {
				curCmd = curValue
				curValue = ""
			} else {
				curValue += string(r)
			}
		default:
			curValue += string(r)
		}
	}
}

func setInventoryData(i *inventory.Inventory) error {
	f, err := os.Open(args.ConfigPath)
	if err != nil {
		return fmt.Errorf("cannot open nginx config file '%s': %w", args.ConfigPath, err)
	}
	defer f.Close()

	err = populateInventory(bufio.NewReader(f), i)
	if err != nil {
		return fmt.Errorf("error parsing inventory from nginx config file '%s': %w", args.ConfigPath, err)
	}

	return nil
}
