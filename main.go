package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/schollz/progressbar"
	"time"
)

func main() {
	fmt.Println("Welcome to the SCPI CLI!")
	fmt.Println("Please use `CTRL-D` to exit this program..")
	defer fmt.Println("Bye!")

	ic := ipCompleter{}
	sc := scpiCompleter{} //TODO: Pass in instrument

	address := prompt.Input(
		"IP Address: ",
		ic.completer,
		prompt.OptionCompletionWordSeparator("."))

	ConnectToInstrument(address)
	sc.prepareScpiCompleter()

	p := prompt.New(
		simpleExecutor,
		sc.completer,
		prompt.OptionTitle("SCPI CLI (SCliPI)"),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(":"))

	p.Run()
}

func ConnectToInstrument(address string) { //TODO: return instrument?
	bar := progressbar.New(1000)
	for i := 0; i < 1000; i++ {
		bar.Add(1)
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println("You selected " + address)
}
