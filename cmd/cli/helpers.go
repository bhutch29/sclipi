package main

import (
	"fmt"
	"log"
	"time"
)

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
	defer inst.Close()

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
	defer inst.Close()

	sm := newScpiManager(inst)
	sm.runScript(file)
}

