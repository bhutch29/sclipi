package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type scpiNode struct {
	Content  nodeInfo
	Children []scpiNode
}

type nodeInfo struct {
	Text     string
	Start    int
	Stop     int
	Suffixed bool
}

func parseScpi(lines []string) scpiNode {
	head := scpiNode{}
	commands := splitScpiCommands(lines)

	for _, command := range commands {
		createScpiTreeBranch(command, &head)
	}

	return head
}

func createScpiTreeBranch(command []nodeInfo, head *scpiNode) {
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
}

func scpiNodeExists(nodes []scpiNode, info nodeInfo) (bool, int) {
	for i, node := range nodes {
		if node.Content == info {
			return true, i
		}
	}

	return false, -1
}

func splitScpiCommands(lines []string) [][]nodeInfo {
	var commands [][]nodeInfo
	for _, line := range lines {
		line = strings.Replace(line, "[", "", -1)
		trimmed := strings.TrimLeft(line, ":")
		suffixed := reformatSuffixes(trimmed)
		split := strings.Split(suffixed, ":")
		withOptionals := handleOptionals(removeSquareBraces(split), getOptionalIndexes(split))
		withQueries := handleQueries(withOptionals)
		withoutBars := handleBars(withQueries)
		finished := finishSuffixes(withoutBars)

		commands = append(commands, finished...)
	}
	return commands
}

func finishSuffixes(commands [][]string) [][]nodeInfo {
	var result [][]nodeInfo
	for _, command := range commands {
		var commandInfo []nodeInfo
		for _, subcommand := range command {
			if atIndex := strings.Index(subcommand, "@"); atIndex != -1 {
				startVal, err := strconv.Atoi(string(subcommand[atIndex+1]))
				if err != nil {
					log.Fatal("Failed to parse one of the available SCPI commands: ", subcommand)
				}
				end := string(subcommand[atIndex+3:])
				stopVal, err := strconv.Atoi(strings.TrimSuffix(end, "?"))
				if err != nil {
					log.Fatal("Failed to parse one of the available SCPI commands: ", subcommand)
				}
				text := subcommand[:atIndex]
				if strings.HasSuffix(subcommand, "?") {
					text += "?"
				}
				commandInfo = append(commandInfo, nodeInfo{Text: text, Suffixed: true, Start: startVal, Stop: stopVal})
			} else {
				commandInfo = append(commandInfo, nodeInfo{Text: subcommand, Suffixed: false})
			}
		}
		result = append(result, commandInfo)
	}
	return result
}

func handleBars(commands [][]string) [][]string {
	var result [][]string
	for _, command := range commands {
		barIndexes := getBarIndexes(command)
		result = append(result, extractBarCommands(command, barIndexes)...)
	}
	return result
}

//Recursively walk "tree" of command depth-first, returning all possible combinations of "bar" commands
func extractBarCommands(command []string, barIndexes []int) [][]string {
	var result [][]string

	if len(barIndexes) == 0 {
		return append(result, command)
	}

	options := strings.Split(command[barIndexes[0]], "|")
	checkForQuerySuffix(options)

	result = append(result, replaceAndRecurse(command, barIndexes, options[0])...)
	result = append(result, replaceAndRecurse(command, barIndexes, options[1])...)

	return result
}

func checkForQuerySuffix(options []string) {
	if strings.HasSuffix(options[0], "?") && !strings.HasSuffix(options[1], "?") {
		options[1] += "?"
	}
	if strings.HasSuffix(options[1], "?") && !strings.HasSuffix(options[0], "?") {
		options[0] += "?"
	}
}

func replaceAndRecurse(command []string, barIndexes []int, option string) [][]string {
	commandCopy := make([]string, len(command))
	copy(commandCopy, command)

	commandCopy[barIndexes[0]] = option
	return extractBarCommands(commandCopy, barIndexes[1:])
}

func getBarIndexes(command []string) []int {
	var indexes []int
	for i, text := range command {
		if strings.Contains(text, "|") {
			indexes = append(indexes, i)
		}
	}
	return indexes
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

// Rewrites any discovered suffixes into an easier to parse format that most importantly doesnt have any ':' characters
func reformatSuffixes(s string) string {
	r, _ := regexp.Compile("{([0-9]):([0-9][0-9]?)}")
	match := r.FindStringSubmatchIndex(s)
	if match == nil {
		return s
	}
	startVal := string(s[match[2]])
	stopVal := calculateStopSuffix(s, match)

	startCut := match[0]
	stopCut := match[1]

	result := s[:startCut] + "@" + string(startVal) + "#" + string(stopVal) + s[stopCut:]
	return reformatSuffixes(result)
}

//Determines whether the stop value of the suffix is double digit or not, then generates the correct value
func calculateStopSuffix(s string, match []int) string {
	if match[5]-match[4] == 1 {
		return string(s[match[4]])
	} else {
		digit1 := string(s[match[4]])
		digit2 := string(s[match[4]+1])
		return digit1 + digit2
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
