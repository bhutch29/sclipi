package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"log"
	"time"
)

var version = "undefined"

func main() {
	args := parseArgs()
	commonPromptOptions := getPromptOptions(args)

	printIntroText(*args.Quiet)
	defer fmt.Println("Goodbye!")
	address := getAddress(args, commonPromptOptions)

	bar := Progress{Silent: *args.Quiet}
	bar.Forward(33)

	inst, err := buildAndConnectInstrument(address, *args.Port)
	if err != nil {
		fmt.Println()
		log.Fatal(err)
	}
	defer inst.close()

	bar.Forward(34)
	sm := newScpiManager(inst)
	bar.Forward(33)

	if !*args.Quiet {
		fmt.Println("Connected!")
	}

	options := []prompt.Option{
		prompt.OptionTitle("Sclipi (SCPI cli)"),
		prompt.OptionCompletionWordSeparator(":"),
	}
	p := prompt.New(sm.executor, sm.completer, append(options, commonPromptOptions...)...)

	p.Run()
}

func getPromptOptions(args arguments) []prompt.Option {
	return []prompt.Option{
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionInputTextColor(args.TextColor),
		prompt.OptionSuggestionTextColor(args.SuggestionColor),
		prompt.OptionSuggestionBGColor(args.SuggestionBgColor),
		prompt.OptionSelectedSuggestionTextColor(args.SelectedColor),
		prompt.OptionSelectedSuggestionBGColor(args.SelectedBgColor),
		prompt.OptionDescriptionTextColor(args.SelectedColor),
		prompt.OptionDescriptionBGColor(args.SelectedBgColor),
		prompt.OptionSelectedDescriptionTextColor(args.SuggestionColor),
		prompt.OptionSelectedDescriptionBGColor(args.SuggestionBgColor),
		prompt.OptionPreviewSuggestionTextColor(args.PreviewColor),
	}
}

func printIntroText(silent bool) {
	if silent {
		return
	}
	fmt.Println("Welcome to the SCPI cli!")
	fmt.Println("Use Tab to navigate auto-completion options")
	fmt.Println("Use `CTRL-D`, `quit`, or `exit` to exit this program")
}

func printHelp() {
	fmt.Println()
	fmt.Println("# Sclipi's tab-completion is operated entirely using the Tab key")
	fmt.Println("#     Press Tab repeatedly to cycle through the available options")
	fmt.Println("#     Typing will filter the list")
	fmt.Println("#     Pressing the Right Arrow key or continuing to type will accept the selected option")
	fmt.Println("#     Up and Down arrow keys cycle through your command history")
	fmt.Println()
}

func getAddress(args arguments, commonOptions []prompt.Option) string {
	if *args.Simulate {
		return "simulated"
	}
	if *args.Address != "" {
		return *args.Address
	}

	ic := newIpCompleter(simFileExists())
	var result string
	for {
		options := []prompt.Option{
			prompt.OptionTitle("Sclipi (SCPI cli)"),
			prompt.OptionCompletionWordSeparator("."),
		}
		result = prompt.Input("Address: ", ic.completer, append(options, commonOptions...)...)
		if result != "?" {
			break
		}
		printHelp()
	}
	return result
}

func buildAndConnectInstrument(address string, port string) (instrument, error) {
	var inst instrument
	if address == "simulated" {
		inst = &simInstrument{}
	} else {
		inst = &scpiInstrument{}
	}

	if err := inst.connect(5*time.Second, address+":"+port); err != nil {
		return inst, err
	}

	return inst, nil
}

type Progress struct {
	Silent      bool
	bar         *progressbar.ProgressBar
	initialized bool
}

func (p *Progress) Forward(percent int) {
	if p.Silent {
		return
	}
	if !p.initialized {
		p.bar = progressbar.New(100)
		p.initialized = true
	}
	_ = p.bar.Add(percent)
}
