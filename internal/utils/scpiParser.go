package utils

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"slices"
)

type ScpiNode struct {
  Content  nodeInfo `json:"content"`
  Children []ScpiNode `json:"children"`
}

type nodeInfo struct {
  Text     string `json:"text"`
  Start    int `json:"start"`
  Stop     int `json:"stop"`
  Suffixed bool `json:"suffixed"`
}

func parseScpi(lines []string) ScpiNode {
	head := ScpiNode{}
	commands := splitScpiCommands(lines)

	for _, command := range commands {
		createScpiTreeBranch(command, &head)
	}

	return head
}

func createScpiTreeBranch(command []nodeInfo, head *ScpiNode) {
	if len(command) == 0 {
		return
	}
	if exists, index := scpiNodeExists(head.Children, command[0]); exists {
    if command[0].Start < head.Children[index].Content.Start {
      head.Children[index].Content.Start = command[0].Start
    }
    if command[0].Stop > head.Children[index].Content.Stop {
      head.Children[index].Content.Stop = command[0].Stop
    }
		createScpiTreeBranch(command[1:], &head.Children[index])
	} else {
		insertIndex := findSortedInsertIndex(head.Children, command[0])
		head.Children = slices.Insert(head.Children, insertIndex, ScpiNode{Content: command[0]})
    if len(command) > 1 {
		  createScpiTreeBranch(command[1:], &head.Children[insertIndex])
    }
	}
}

func findSortedInsertIndex(nodes []ScpiNode, info nodeInfo) int {
	for i, node := range nodes {
		if info.Text < node.Content.Text {
			return i
		}
		if info.Text == node.Content.Text && !info.Suffixed && node.Content.Suffixed {
			return i
		}
	}
	return len(nodes)
}

func scpiNodeExists(nodes []ScpiNode, info nodeInfo) (bool, int) {
	for i, node := range nodes {
		if node.Content.Text == info.Text && node.Content.Suffixed == info.Suffixed {
			return true, i
		}
	}

	return false, -1
}

// Converts :SYSTem:HELP:HEADers?-style SCPI definitions into a complete list of possible SCPI commands/queries.
// NodeInfo objects contain command "suffix" information, e.g. RADio{1:16}
func splitScpiCommands(lines []string) [][]nodeInfo {
	var commands [][]nodeInfo
	for _, line := range lines {
		s := strings.Replace(line, "[", "", -1)
		s = strings.TrimLeft(s, ":")
		//TODO: Suffixed items are also accessible without suffix (default value)
		s = reformatSuffixes(s)
		s = reformatIrregularSuffixes(s)
		//TODO: Convert all methods up to finishSuffixes to use strings instead of slices, will enable speeding up finishSuffixes by switching it from loops to recursion.
		ss := strings.Split(s, ":")
		sss := handleOptionals(removeSquareBraces(ss), getOptionalIndexes(ss))
		sss = handleQueries(sss)
		sss = handleBars(sss)
		nodeInfos := finishSuffixes(sss)

		commands = append(commands, nodeInfos...)
	}
	return commands
}

// Rewrites any discovered suffixes into an easier to parse format that most importantly doesnt have any ':' characters
func reformatSuffixes(s string) string {
	// First handle {N:M} format
	r, _ := regexp.Compile("{([0-9]):([0-9][0-9]?)}")
	match := r.FindStringSubmatchIndex(s)
	if match != nil {
		start := calculateSuffix(s, match[2], match[3])
		stop := calculateSuffix(s, match[4], match[5])

		startCut := match[0]
		stopCut := match[1]

		return reformatSuffixes(s[:startCut] + "@" + start + "#" + stop + s[stopCut:])
	}

	// Handle {N} format as {N:N}
	r2, _ := regexp.Compile("{([0-9][0-9]?)}")
	match2 := r2.FindStringSubmatchIndex(s)
	if match2 == nil {
		return s
	}
	value := calculateSuffix(s, match2[2], match2[3])

	startCut := match2[0]
	stopCut := match2[1]

	return reformatSuffixes(s[:startCut] + "@" + value + "#" + value + s[stopCut:])
}

// Workaround for MXG SCPI existence of RAD1 and RAD{1:1} syntax simultaneously
func reformatIrregularSuffixes(s string) string {
	r, _ := regexp.Compile(`[^\d#@](\d{1,2}):`) // All 1 or 2 digit numbers just before ':' but not preceded by a digit, a #, or a @
	match := r.FindStringSubmatchIndex(s)
	if match == nil {
		return s
	}
	startCut := match[2]
	stopCut := match[3]
	value := s[startCut:stopCut]

	return reformatIrregularSuffixes(s[:startCut] + "@" + value + "#" + value + s[stopCut:])
}

// Finds the temporarily adjusted suffix information and moves it into nodeInfo fields
func finishSuffixes(commands [][]string) [][]nodeInfo {
	//TODO: Need to speed this up
	var result [][]nodeInfo
	for _, command := range commands {
		var commandInfo []nodeInfo
		for _, subcommand := range command {
			commandInfo = append(commandInfo, finishSuffix(subcommand))
		}
		result = append(result, commandInfo)
	}
	return result
}

func finishSuffix(subcommand string) nodeInfo {
	r, _ := regexp.Compile(`@(\d{1,2})#(\d{1,2})`)
	match := r.FindStringSubmatchIndex(subcommand)
	if match == nil {
		return nodeInfo{Text: subcommand, Suffixed: false}
	}

	start := tryConvertAtoi(calculateSuffix(subcommand, match[2], match[3]))
	stop := tryConvertAtoi(calculateSuffix(subcommand, match[4], match[5]))

	startCut := match[0]
	text := subcommand[:startCut]
	if strings.HasSuffix(subcommand, "?") {
		text += "?"
	}
	return nodeInfo{Text: text, Suffixed: true, Start: start, Stop: stop}
}

func tryConvertAtoi(character string) int {
	result, err := strconv.Atoi(character)
	if err != nil {
		log.Fatal("Failed to parse suffix from SCPI command: ")
	}
	return result
}

//Determines whether the value is double digit or not, then generates the correct value string
func calculateSuffix(s string, startIndex int, stopIndex int) string {
	if stopIndex-startIndex == 1 {
		return string(s[startIndex])
	} else {
		digit1 := string(s[startIndex])
		digit2 := string(s[startIndex+1])
		return digit1 + digit2
	}
}

func handleBars(commands [][]string) [][]string {
	var result [][]string
	for _, command := range commands {
		result = append(result, extractBarCommands(command, getBarIndexes(command))...)
	}
	return result
}

//Recursively walk "tree" of command depth-first, returning all possible combinations of "bar" commands, e.g. Option1|Option2
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

// Converts the various command types: qonly (query-only), nquery(no query), and default (command and query)
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
			query[last] += "?"
			result = append(result, query)
		}
	}
	return result
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

func deleteIndexFromSliceRetainingQueryInfo(command []string, index int) []string {
	newCommand := make([]string, len(command))
	copy(newCommand, command)
	if index == len(newCommand)-1 {
		if strings.HasSuffix(newCommand[index], "/nquery/") {
			newCommand[index-1] += "/nquery/"
		} else if strings.HasSuffix(command[index], "?/qonly/") {
			newCommand[index-1] += "?/qonly/"
		}
	}
	return slices.Delete(newCommand, index, index+1)
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
