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

			for _, command := range withOptionals {
				commands = append(commands, command)
			}
		}
	}
	return commands
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
		cut := append(temp[:startCut], s[stopCut - 1:]...)
		cut[startCut] = byte(i)
		//result := string(cut) + string(i)
		suffixes = append(suffixes, branchSuffixes(string(cut))...)
	}

	return suffixes
}

func isSuffixed(s string) bool {
	return strings.Contains(s, "{")
}

func generateOptionalCommands(command []string, optionalIndexes []int) [][]string {
	commands := [][]string{removeSquareBraces(command)}
	for i, index := range optionalIndexes {
		shortened := deleteStringFromSlice(command, index) //TODO: if last element is optional, deletes query information
		remaining := optionalIndexes[i+1:]
		commands = append(commands, generateOptionalCommands(shortened, remaining)...)
	}
	return commands
}

func deleteStringFromSlice(split []string, index int) []string {
	return append(split[:index], split[index+1:]...)
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
