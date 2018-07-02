package scpiParser

import (
	"strings"
	"regexp"
	"strconv"
	"fmt"
)

func splitScpiCommands(lines []string) [][]string {
	var commands [][]string
	for _, line := range lines {
		line = strings.Replace(line, "[", "", -1)
		trimmed := strings.TrimLeft(line, ":")
		split := strings.Split(trimmed, ":")
		withOptionals := generateOptionalCommands(split, getOptionalIndexes(split))

		var withSuffixes [][]string
		for _, command := range withOptionals {
			withSuffixes = append(withSuffixes, generateSuffixedCommands(command)...)
		}

		for _, command := range withSuffixes {
			commands = append(commands, command)
		}
	}
	return commands
}
func generateSuffixedCommands(command []string) [][]string {
	var commands [][]string
	suffixed := false
	for i, word := range command{
		if isSuffixed(word) {
			suffixed = true
			for _, suffix := range branchSuffixes(word){
				temp := command
				temp[i] = suffix
				generateSuffixedCommands(temp)
			}
		}
	}

	if suffixed {
		return [][]string{}
	} else {
		return commands
	}

}
func branchSuffixes(s string) []string{
	var suffixes []string
	r, _ := regexp.Compile("([^{]*){([0-9]):([0-9])}")
	result := r.FindStringSubmatch(s)
	prefix := result[1]
	start, err := strconv.Atoi(result[2])
	if err != nil {
		fmt.Println("Error parsing start value of suffix, e.g. {1:2}. Error = " + err.Error())
		return []string{}
	}
	stop, err := strconv.Atoi(result[3])
	if err != nil {
		fmt.Println("Error parsing stop value of suffix, e.g. {1:2}. Error = " + err.Error())
		return []string{}
	}

	for i := start; i <= stop; i++  {
		suffixes = append(suffixes, prefix + strconv.Itoa(i))
	}

	return suffixes
}

func isSuffixed(s string) bool{
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
