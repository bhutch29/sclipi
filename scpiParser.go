package main

import (
	"strings"
	"regexp"
	"os"
	"bufio"
)

type scpiNode struct {
	Content  string
	Children []scpiNode
}

func parseScpi(lines []string) scpiNode {
	head := scpiNode{}
	commands := splitScpiCommands(lines)

	for _, command := range commands {
		createScpiTreeBranch(command, &head)
	}

	return head
}

func createScpiTreeBranch(command []string, head *scpiNode) {
	if len(command) == 0 {
		return
	}
	if exists, index := scpiNodeExists(head.Children, command[0]); exists {
		if len(command) == 1 {
			return
		} else {
			createScpiTreeBranch(command[1:], &head.Children[index])
		}
	} else {
		head.Children = append(head.Children, scpiNode{Content: command[0]})
		if len(command) > 1 {
			createScpiTreeBranch(command[1:], &head.Children[len(head.Children) - 1])
		}
	}
	return
}

func scpiNodeExists(nodes []scpiNode, word string) (bool, int) {
	for i, node := range nodes{
		if node.Content == word {
			return true, i
		}
	}

	return false, -1
}

func splitScpiCommands(lines []string) [][]string {
	var commands [][]string
	for _, line := range lines {
		line = strings.Replace(line, "[", "", -1)
		trimmed := strings.TrimLeft(line, ":")
		suffixed := branchSuffixes(trimmed)

		for _, item := range suffixed {
			split := strings.Split(item, ":")
			withOptionals := generateOptionalCommands(removeSquareBraces(split), getOptionalIndexes(split))
			withQueries := generateQueryCommands(withOptionals)

			for _, command := range withQueries {
				commands = append(commands, command)
			}
		}
	}
	return commands
}
func generateQueryCommands(commands [][]string) [][]string {
	var result [][]string
	for _, command := range commands {
		last := len(command) - 1
		if strings.Contains(command[last], "?/qonly/") {
			command[last] = strings.Replace(command[last], "?/qonly/", "?", -1)
			result = append(result, command)
		} else if strings.Contains(command[last], "/nquery/") {
			command[last] = strings.Replace(command[last], "/nquery/", "", -1)
			result = append(result, command)
		} else {
			result = append(result, command)
			query := make([]string, len(command))
			copy(query, command)
			query[last] = query[last] + "?"
			result = append(result, query)
		}
	}
	return result
}

func branchSuffixes(s string) []string {
	var suffixes []string
	r, _ := regexp.Compile("{([0-9]):([0-9])}")
	match := r.FindStringSubmatchIndex(s)
	if match == nil {
		return []string{s}
	}
	startSuffix := s[match[2]]
	stopSuffix := s[match[4]]
	startCut := match[0]
	stopCut := match[1]

	for i := startSuffix; i <= stopSuffix; i++ {
		temp := []byte(s)
		cut := append(temp[:startCut], s[stopCut-1:]...) //Cut suffix portion out, leave one element for value
		cut[startCut] = byte(i)                          //Insert value into spare element
		suffixes = append(suffixes, branchSuffixes(string(cut))...)
	}

	return suffixes
}

func generateOptionalCommands(command []string, optionalIndexes []int) [][]string {
	commands := [][]string{command}
	for i, index := range optionalIndexes {
		shortened := deleteIndexFromSliceRetainingQueryInfo(command, index)
		remaining := removeIndexAndDecrement(optionalIndexes, i)
		commands = append(commands, generateOptionalCommands(shortened, remaining)...)
	}
	return commands
}

func removeIndexAndDecrement(indexes []int, i int) []int {
	newIndexes := make([]int, len(indexes)) //TODO: Understand better why this is necessary
	copy(newIndexes, indexes)
	for j := range newIndexes {
		if j > i {
			newIndexes[j]--
		}
	}
	return newIndexes[i+1:]
}

func deleteIndexFromSliceRetainingQueryInfo(command []string, index int) []string {
	newCommand := make([]string, len(command))
	copy(newCommand, command)
	if index == len(newCommand)-1 && strings.Contains(newCommand[index], "/") {
		if strings.Contains(newCommand[index], "/nquery/") {
			newCommand[index-1] = newCommand[index-1] + "/nquery/"
		} else if strings.Contains(command[index], "?/qonly/") {
			newCommand[index-1] = newCommand[index-1] + "?/qonly/"
		}
	}
	return append(newCommand[:index], newCommand[index+1:]...)
}

func getOptionalIndexes(words []string) []int {
	var optionals []int
	for i, item := range words {
		if strings.Contains(item, "]") {
			optionals = append(optionals, i)
		}
	}
	return optionals
}

func removeSquareBraces(words []string) []string {
	var result []string
	for _, word := range words {
		result = append(result, strings.Replace(word, "]", "", -1))
	}
	return result
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