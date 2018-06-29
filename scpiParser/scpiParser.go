package scpiParser

import "strings"

func splitScpiCommands(lines []string) [][]string {
	var commands [][]string
	for _, line := range lines {
		line = strings.Replace(line,"[", "", -1)
		trimmed := strings.TrimLeft(line, ":")
		split := strings.Split(trimmed, ":")

		var includingOptionals [][]string
		includingOptionals = append(includingOptionals, removeSquareBraces(split)) //Add case where everything is used

		optionals := getOptionals(split)
		for _, index := range optionals{ //TODO: Only handles case with single optional, need to recurse
			var a []string
			a = append(split[:index], split[index+1:]...) //delete element of optional index from slice
			includingOptionals = append(includingOptionals, removeSquareBraces(a))
		}

		//TODO: add generated list to commands variable
	}
	return commands
}

func getOptionals(words []string) []int{
	var optionals []int
	for i, item := range words{
		if strings.Contains(item,"]") {
			optionals = append(optionals, i)
		}
	}
	return optionals
}

func removeSquareBraces(words []string) []string{
	var result []string
	for _, word := range words{
		result = append(result, strings.Replace(word, "]", "", -1))
	}
	return result
}