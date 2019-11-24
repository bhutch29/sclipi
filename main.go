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
	scriptFileFlag := parser.String("f", "file", &argparse.Options{Help: "The path to a newline-delimited list of commands to be run non-interactively. Must set IP address if using this feature"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	port := parsePort(*portFlag)
	if *scriptFileFlag != "" {
		runScriptFile(*scriptFileFlag, *ipFlag, port)
		return
	}

	fmt.Println("Welcome to the SCPI CLI!")
	fmt.Println("Please use `CTRL-D` or `quit` to exit this program..")
	defer fmt.Println("Goodbye!")

	address := getAddress(*ipFlag)

	bar := progressbar.New(100)
	_ = bar.Add(25)

	inst, err := buildAndConnectInstrument(address, port)
	if err != nil {
		fmt.Println()
		log.Fatal(err)
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

func parsePort(port string) string {
	if port == "" {
		port = "5025"
	}
	return port
}

func runScriptFile(file string, ip string, port string) {
		if ip == "" {
			log.Fatal("Error: IP flag must be set when using File flag")
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

func getAddress(ip string) string {
	if ip == "" {
		ic := ipCompleter{}
		return prompt.Input(
			"IP Address: ",
			ic.completer,
			prompt.OptionCompletionWordSeparator("."))
	}
	return ip
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
