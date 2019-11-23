package main

import (
	"fmt"
	"os"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"github.com/akamensky/argparse"
	"log"
	"time"
)

func main() {
	parser := argparse.NewParser("SCliPI", "A SCPI CLI. Features autocomplete and much more")
	ipFlag := parser.String("i", "ip", &argparse.Options{Help: "The IP address of the instrument. If not provided, SCliPI will use your network information and auto-completion to assist you"})
	portFlag := parser.String("p", "port", &argparse.Options{Help: "The SCPI port of the instrument. Defaults to 5025"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	fmt.Println("Welcome to the SCPI CLI!")
	fmt.Println("Please use `CTRL-D` or `quit` to exit this program..")
	defer fmt.Println("Goodbye!")

	var address string
	if *ipFlag == "" {
		ic := ipCompleter{}
		address = prompt.Input(
			"IP Address: ",
			ic.completer,
			prompt.OptionCompletionWordSeparator("."))
	} else {
		address = *ipFlag
	}

	var port string
	if *portFlag == "" {
		port = "5025"
	} else {
		port = *portFlag
	}

	bar := progressbar.New(100)
	_ = bar.Add(25)

	inst, err := buildAndConnectInstrument(address, port)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		log.Fatal()
	}
	defer inst.Close()

	_ = bar.Add(25)
	sm := newScpiManager(inst)
	_ = bar.Add(50)

	fmt.Println("Connected!")

	p := prompt.New(
		sm.executor,
		sm.completer,
		prompt.OptionTitle("SCPI CLI (SCliPI)"),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(":"))

	p.Run()
}

func buildAndConnectInstrument(address string, port string) (instrument, error) {
	var inst instrument
	if address == "simulated" {
		inst = &simInstrument{}
	} else {
		inst = &scpiInstrument{}
	}

	if err := inst.Connect(5 * time.Second, address + ":" + port); err != nil {
		return inst, err
	}

	return inst, nil
}
