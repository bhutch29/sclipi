package scpiParser

import (
	"testing"
	"os"
	"bufio"
)

func TestScpiParser(t *testing.T) {
	lines := []string {":ABORt/nquery/"}
	commands := splitScpiCommands(lines)
	if len(commands) != 1 {
		t.Error(":ABORt/nquery/ not parsed properly:", commands[0])
		return
	}
	if len(commands[0]) != 1 {
		t.Error(":ABORt/nquery/ not parsed properly:", commands[0])
	}

	lines = []string {":ABORt[:SWEep]/nquery/"}
	commands = splitScpiCommands(lines)
	if len(commands) != 2 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 1 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands[0])
	}
	if len(commands[1]) != 2 {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands[1])
	}

	lines = []string {":CALibration:BBG:CHANnel:OFFSet"}
	commands = splitScpiCommands(lines)
	if len(commands) != 2 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands)
		return
	}
	if len(commands[0]) != 4 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands[0])
	}
	if len(commands[1]) != 4 {
		t.Error(":CALibration:BBG:CHANnel:OFFSet not parsed properly:", commands[1])
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}