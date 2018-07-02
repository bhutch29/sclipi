package scpiParser

import (
	"strings"
	"regexp"
)

func splitScpiCommands(lines []string) [][]string {
	var commands [][]string
	for _, line := range lines {
		line = strings.Replace(line, "[", "", -1)
		trimmed := strings.TrimLeft(line, ":")
		suffixed := branchSuffixes(trimmed)

		for _, item := range suffixed {
			split := strings.Split(item, ":")
			withOptionals := generateOptionalCommands(split, getOptionalIndexes(split))
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
	for _, command := range commands{
		last := len(command) - 1
		if strings.Contains(command[last], "?/qonly/"){
			command[last] = strings.Replace(command[last],"?/qonly/", "?", -1)
			result = append(result, command)
		} else if strings.Contains(command[last], "/nquery/"){
			command[last] = strings.Replace(command[last],"/nquery/", "", -1)
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
		cut := append(temp[:startCut], s[stopCut - 1:]...) //Cut suffix portion out, leave one element for value
		cut[startCut] = byte(i) //Insert value into spare element
		suffixes = append(suffixes, branchSuffixes(string(cut))...)
	}

	return suffixes
}

func generateOptionalCommands(command []string, optionalIndexes []int) [][]string {
	commands := [][]string{removeSquareBraces(command)}
	for i, index := range optionalIndexes {
		shortened := deleteIndexFromSliceRetainingQueryInfo(command, index) //TODO: if last element is optional, deletes query information
		remaining := optionalIndexes[i+1:]
		commands = append(commands, generateOptionalCommands(shortened, remaining)...)
	}
	return commands
}

func deleteIndexFromSliceRetainingQueryInfo(command []string, index int) []string {
	if index == len(command)-1 && strings.Contains(command[index], "/"){
		if strings.Contains(command[index], "/nquery/"){
			command[index - 1] = command[index - 1] + "/nquery/"
		} else if strings.Contains(command[index], "?/qonly/"){
			command[index - 1] = command[index - 1] + "?/qonly/"
		}
	}
	return append(command[:index], command[index+1:]...)
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
