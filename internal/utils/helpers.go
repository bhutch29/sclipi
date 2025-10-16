package utils

import (
	"bufio"
	"github.com/shibukawa/configdir"
	"os"
)

func ReadLinesFromFile(file *os.File) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func ReadLinesFromPath(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadLinesFromFile(file)
}

func SimFileExists() bool {
	info, err := os.Stat("SCPI.txt")
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// func writeCommandsToFile(commands [][]nodeInfo) {
// 	//Only writes the non-star commands to file
// 	if !strings.HasPrefix(commands[0][0].Text, "*") {
// 		f, _ := os.Create("temp.txt")
// 		for _, command := range commands {
// 			for _, subcommand := range command {
// 				fmt.Fprint(f, subcommand.Text+":")
// 			}
// 			fmt.Fprint(f, "\n")
// 		}
// 	}
// }

func GetHistoryFromFile() ([]string, *configdir.Config) {
	var entries []string
	configDirs := configdir.New("bhutch29", "sclipi")
	cache := configDirs.QueryCacheFolder()
	if cache.Exists("history.txt") {
		file, _ := cache.Open("history.txt")
		commands, err := ReadLinesFromFile(file)
		if err == nil {
			for _, command := range commands {
				entries = append(entries, command)
			}
		}
	}
	return entries, cache
}
