package main

import (
	"fmt"
	"github.com/bhutch29/sclipi/internal/utils"
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
	defer inst.Close()

	bar.forward(30)
	sm := newScpiManager(inst)
	bar.forward(30)

	if !*args.Quiet {
		bar.clear()
	}

	history, _ := utils.GetHistoryFromFile()

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
	fmt.Println("Use `CTRL-D` or `quit` to exit this program")
}

func printHelp() {
	fmt.Println();
	fmt.Println(`# Basics:
Sclipi is an application for sending SCPI commands and receiving responses.
SCPI commands start with a colon, type ':' to get started.

# Completion:
Completion options appear automatically. Typing will filter the list of options.
Completion is initiated with either the Tab or Down arrow keys.
Press Tab or Down repeatedly to cycle through the available options.
Shift-Tab and Up will cycle back up through the options.
Pressing the Right Arrow key or continuing to type will accept the selected option.
Just because a completion doesn't appear doesn't mean your command won't work! Some commands are hidden.

# History:
Sclipi tracks the history of all commands you have ever sent.
Up and Down arrow keys cycle through your command history.

# Exiting:
There are 3 ways to exit the application.
1. Type 'quit' and hit Enter
2. Type 'exit' and hit Enter
3. Clear the input text and press Ctrl-D`)
	fmt.Println();
}

func getAddress(args arguments, commonOptions []prompt.Option) string {
	if *args.Simulate {
		return "simulated"
	}
	if *args.Address != "" {
		return *args.Address
	}

	ic := newIpCompleter(utils.SimFileExists())
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

func buildAndConnectInstrument(address string, port string, timeout time.Duration, bar *progress) (utils.Instrument, error) {
	var inst utils.Instrument
	if address == "simulated" {
		inst = utils.NewSimInstrument(timeout)
	} else {
		inst = utils.NewScpiInstrument(timeout)
	}

	if err := inst.Connect(address+":"+port, bar.forward); err != nil {
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

func (p *progress) clear() {
	if p.Silent || !p.initialized {
		return
	}
	// Clear the progress bar line by moving cursor to start and clearing the line
	fmt.Print("\r\033[K")
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
