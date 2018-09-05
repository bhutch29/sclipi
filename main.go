package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"log"
)

func main() {
	fmt.Println("Welcome to the SCPI CLI!")
	fmt.Println("Please use `CTRL-D` or `quit` to exit this program..")
	defer fmt.Println("Bye!")

	ic := ipCompleter{}

	address := prompt.Input(
		"IP Address: ",
		ic.completer,
		prompt.OptionCompletionWordSeparator("."))

	bar := progressbar.New(100)
	bar.Add(25)

	inst, err := buildAndConnectInstrument(address)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}

	bar.Add(25)
	sm := newScpiManager(inst)
	bar.Add(50)

	fmt.Println()

	p := prompt.New(
		sm.executor,
		sm.completer,
		prompt.OptionTitle("SCPI CLI (SCliPI)"),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(":"))

	p.Run()
}

func buildAndConnectInstrument(address string) (instrument, error) {
	var inst instrument
	if address == "simulated" {
		inst = &simInstrument{}
	} else {
		inst = &scpiInstrument{}
	}

	if err := inst.Connect(address + ":5025"); err != nil {
		return inst, err
	}

	return inst, nil
}
