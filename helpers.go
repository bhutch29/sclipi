package main

import (
	"os"
	"fmt"
	"bufio"
	"log"
)

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

func runCommand(command string, ip string, port string) {
	if ip == "" {
		log.Fatal("Error: Address flag must be set when using Command flag")
	}
	inst, err := buildAndConnectInstrument(ip, port)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}
	defer inst.Close()

	sm := newScpiManager(inst)
	sm.handleScpi(command)
}

func runScriptFile(file string, ip string, port string) {
	if ip == "" {
		log.Fatal("Error: Address flag must be set when using File flag")
	}
	inst, err := buildAndConnectInstrument(ip, port)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}
	defer inst.Close()

	sm := newScpiManager(inst)
	sm.runScript(file)
}