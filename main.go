package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"log"
	"time"
)

func main() {
	args := ParseArgs()

	if args.ScriptFile != "" {
		runScriptFile(args.ScriptFile, args.Ip, args.Port)
		return
	}

	fmt.Println("Welcome to the SCPI CLI!")
	fmt.Println("Please use `CTRL-D` or `quit` to exit this program..")
	defer fmt.Println("Goodbye!")

	address := getAddress(args.Ip)

	bar := progressbar.New(100)
	_ = bar.Add(25)

	inst, err := buildAndConnectInstrument(address, args.Port)
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
		prompt.OptionCompletionWordSeparator(":"),
		prompt.OptionInputTextColor(args.TextColor),
		prompt.OptionSuggestionTextColor(args.SuggestionColor),
		prompt.OptionSuggestionBGColor(args.SuggestionBgColor),
		prompt.OptionSelectedSuggestionTextColor(args.SelectedColor),
		prompt.OptionSelectedSuggestionBGColor(args.SelectedBgColor),
		prompt.OptionDescriptionTextColor(args.SelectedColor),
		prompt.OptionDescriptionBGColor(args.SelectedBgColor),
		prompt.OptionSelectedDescriptionTextColor(args.SuggestionColor),
		prompt.OptionSelectedDescriptionBGColor(args.SuggestionBgColor),
		prompt.OptionPreviewSuggestionTextColor(args.PreviewColor))

	p.Run()
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
