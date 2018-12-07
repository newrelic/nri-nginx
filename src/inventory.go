package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/sdk"
)

func populateInventory(reader *bufio.Reader, inventory sdk.Inventory) (err error) {
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
			inventory.SetItem(strings.Join(prefix, "/"), "value", curValue)
			prefix = prefix[:len(prefix)-1]

			curValue = ""
			curCmd = ""
		case '\n':
			// parse end line and ignore spaces
			for r == '\n' || r == ' ' || r == '\t' {
				r, _, _ = reader.ReadRune()
			}
			lineNo++
			err = reader.UnreadRune()
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

		if (err != nil) && (err != io.EOF) {
			break
		}
	}

	return err
}

func setInventoryData(inventory sdk.Inventory) (err error) {
	var f *os.File
	f, err = os.Open(args.ConfigPath)
	if err != nil {
		return err
	}
	defer func() {
		e := f.Close()
		if (e != nil) && (err == nil) { // Don't mask a previous err
			err = e
		}
	}()

	err = populateInventory(bufio.NewReader(f), inventory)
	return err
}
