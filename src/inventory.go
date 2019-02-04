package main

import (
	"bufio"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"io"
	"os"
	"strings"
)

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
			return fmt.Errorf("error occured while checking inventory from nginx config file, error: %v", err)
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
				return fmt.Errorf("Error parsing config file in Line %d", lineNo)
			}
			prefix = prefix[:closeIdx]
		case ';':
			// parse end statement
			prefix = append(prefix, curCmd)
			i.SetItem(strings.Join(prefix, "/"), "value", curValue)
			prefix = prefix[:len(prefix)-1]

			curValue = ""
			curCmd = ""
		case '\n':
			// parse end line and ignore spaces
			for r == '\n' || r == ' ' || r == '\t' {
				r, _, _ = reader.ReadRune()
			}
			lineNo++
			reader.UnreadRune()
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
		return err
	}
	defer f.Close()

	return populateInventory(bufio.NewReader(f), i)
}
