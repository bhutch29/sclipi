package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"strconv"
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
			createScpiTreeBranch(command[1:], &head.Children[len(head.Children)-1])
		}
	}
	return
}

func scpiNodeExists(nodes []scpiNode, word string) (bool, int) {
	for i, node := range nodes {
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
		suffixed := handleSuffixes(trimmed)

		for _, item := range suffixed {
			split := strings.Split(item, ":")
			withOptionals := handleOptionals(removeSquareBraces(split), getOptionalIndexes(split))
			withQueries := handleQueries(withOptionals)

			for _, command := range withQueries {
				commands = append(commands, command)
			}
		}
	}
	return commands
}

func handleQueries(commands [][]string) [][]string {
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

func handleSuffixes(s string) []string {
	var suffixes []string
	r, _ := regexp.Compile("{([0-9]):([0-9][0-9]?)}")
	match := r.FindStringSubmatchIndex(s)
	if match == nil {
		return []string{s}
	}
	startSuffix, _ := strconv.Atoi(string(s[match[2]]))
	stopSuffix := calculateStopSuffix(s, match)
	startCut := match[0]
	stopCut := match[1]

	for i := startSuffix; i <= stopSuffix; i++ {
		var cut []byte
		temp := []byte(s)
		if i < 10 {
			cut = append(temp[:startCut], s[stopCut-1:]...) //Cut suffix portion out, leave one element for value
			cut[startCut] = strconv.Itoa(i)[0] //Insert value into spare element
		} else {
			cut = append(temp[:startCut], s[stopCut-2:]...) //Cut suffix portion out, leave two elements for value
			cut[startCut] = strconv.Itoa(i)[0] //Insert values into spare elements
			test := strconv.Itoa(i)
			cut[startCut + 1] = test[1]
		}
		suffixes = append(suffixes, handleSuffixes(string(cut))...)
	}

	return suffixes
}

//Determines whether the stop value of the suffix is double digit or not, then generates the correct value
func calculateStopSuffix(s string, match []int) int {
	if match[5] - match[4] == 1 {
		result, _ := strconv.Atoi(string(s[match[4]]))
		return result
	} else {
		digit1 := string(s[match[4]])
		digit2 := string(s[match[4] + 1])
		result, _ := strconv.Atoi(digit1 + digit2)
		return result
	}

}

func handleOptionals(command []string, optionalIndexes []int) [][]string {
	commands := [][]string{command}
	for i, index := range optionalIndexes {
		shortened := deleteIndexFromSliceRetainingQueryInfo(command, index)
		remaining := removeIndexAndDecrement(optionalIndexes, i)
		commands = append(commands, handleOptionals(shortened, remaining)...)
	}
	return commands
}

func removeIndexAndDecrement(indexes []int, i int) []int {
	newIndexes := make([]int, len(indexes))
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
