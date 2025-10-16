package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"time"
	"os"
)

var version = "undefined"

func main() {
	args := parseArgs()
	commonPromptOptions := getPromptOptions(args)

	printIntroText(*args.Quiet)
	defer fmt.Println("Goodbye!")
	address := getAddress(args, commonPromptOptions)

	bar := progress{Silent: *args.Quiet}
	bar.forward(0)

	inst, err := buildAndConnectInstrument(address, *args.Port, time.Duration(*args.Timeout) * time.Second, &bar)
	if err != nil {
		fmt.Println()
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer inst.close()

	bar.forward(30)
	sm := newScpiManager(inst)
	bar.forward(30)

	if !*args.Quiet {
		fmt.Println("Connected!")
	}

	history, _ := getHistoryFromFile()

	options := []prompt.Option{
		prompt.OptionTitle("Sclipi (SCPI cli)"),
		prompt.OptionCompletionWordSeparator(":"),
		prompt.OptionHistory(history),
	}
	p := prompt.New(sm.executor, sm.completer, append(options, commonPromptOptions...)...)

	p.Run()
}

func getPromptOptions(args arguments) []prompt.Option {
	return []prompt.Option{
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionCompletionOnDown(),
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

func buildAndConnectInstrument(address string, port string, timeout time.Duration, bar *progress) (instrument, error) {
	var inst instrument
	if address == "simulated" {
		inst = &simInstrument{}
	} else {
		inst = &scpiInstrument{timeout: timeout}
	}

	if err := inst.connect(address+":"+port, bar); err != nil {
		return inst, err
	}

	return inst, nil
}

type progress struct {
	Silent      bool
	bar         *progressbar.ProgressBar
	initialized bool
	add         chan int
	done        chan bool
	progress    int
}

func (p *progress) forward(percent int) {
	if p.Silent {
		return
	}
	if !p.initialized {
		p.progress = 0
		p.done = make(chan bool)
		p.add = make(chan int, 100)
		go p.runBar(p.add)
		p.initialized = true
	}
	p.progress += percent
	p.add <- percent
	if p.progress >= 100 {
		p.close()
	}
}

func (p *progress) close() {
	close(p.add)
	<-p.done
}

func (p *progress) runBar(in <-chan int) {
	p.bar = progressbar.New(100)
	added := 0
	for {
		select {
		case percent, more := <-in:
			if more {
				_ = p.bar.Add(percent - added)
				added = 0
			} else {
				p.done <- true
				return
			}
		case <-time.After(time.Second / 2):
			if added > 20 {
				return
			}
			added++
			_ = p.bar.Add(1)
		}
	}
}
