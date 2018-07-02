package scpiParser

import (
	"testing"
	"os"
	"bufio"
)

func TestScpiParser(t *testing.T) {
	lines := []string {":Example{1:2}:Afterward"}
	commands := splitScpiCommands(lines)
	if len(commands) != 4 {
		t.Error(":Example{1:2}:Afterward not parsed properly:", commands[0])
		return
	}
	if len(commands[0]) != 2 {
		t.Error(":Example{1:2}:Afterward not parsed properly:", commands[0])
	}

	lines = []string {":ABORt/nquery/"}
	commands = splitScpiCommands(lines)
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
	if len(commands[0]) != 2  {
		t.Error(":ABORt[:SWEep]/nquery/ not parsed properly:", commands[0])
	}
	if len(commands[1]) != 1 {
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

func TestBranchSuffixes(t *testing.T) {
	result := branchSuffixes("Example{1:1}")
	if len(result) != 1 {
		t.Error("Example{1:1} not parsed to 1 result", result)
	}

	result = branchSuffixes("Example{1:3}")
	if len(result) != 3 {
		t.Error("Example{1:3} not parsed to 3 results", result)
	}
	if result[0] != "Example1"{
		t.Error("Example{1:3} first element not Example1", result)
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