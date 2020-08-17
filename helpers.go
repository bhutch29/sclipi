package main

import (
	"bufio"
	"fmt"
	"github.com/shibukawa/configdir"
	"log"
	"os"
	"strings"
	"time"
)

func readLinesFromFile(file *os.File) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func readLinesFromPath(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readLinesFromFile(file)
}

func runCommand(command string, ip string, port string, timeout time.Duration) {
	if ip == "" {
		log.Fatal("Error: Address flag must be set when using Command flag")
	}
	inst, err := buildAndConnectInstrument(ip, port, timeout, &progress{})
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}
	defer inst.close()

	sm := newScpiManager(inst)
	sm.handleScpi(command)
}

func runScriptFile(file string, ip string, port string, timeout time.Duration) {
	if ip == "" {
		log.Fatal("Error: Address flag must be set when using File flag")
	}
	inst, err := buildAndConnectInstrument(ip, port, timeout, &progress{})
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}
	defer inst.close()

	sm := newScpiManager(inst)
	sm.runScript(file)
}

func simFileExists() bool {
	info, err := os.Stat("SCPI.txt")
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeCommandsToFile(commands [][]nodeInfo) {
	//Only writes the non-star commands to file
	if !strings.HasPrefix(commands[0][0].Text, "*") {
		f, _ := os.Create("temp.txt")
		for _, command := range commands {
			for _, subcommand := range command {
				fmt.Fprint(f, subcommand.Text+":")
			}
			fmt.Fprint(f, "\n")
		}
	}
}

func getHistoryFromFile() ([]string, *configdir.Config) {
	var entries []string
	configDirs := configdir.New("bhutch29", "sclipi")
	cache := configDirs.QueryCacheFolder()
	if cache.Exists("history.txt") {
		file, _ := cache.Open("history.txt")
		commands, err := readLinesFromFile(file)
		if err == nil {
			for _, command := range commands {
				entries = append(entries, command)
			}
		}
	}
	return entries, cache
}