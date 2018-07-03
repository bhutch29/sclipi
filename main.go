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

	address := prompt.Input("IP Address: ", ipCompleter, prompt.OptionCompletionWordSeparator("."))
	ConnectToInstrument(address)

	prepareScpiCompleter() //TODO: refactor. This should speed up first SCPI command though

	p := prompt.New(
		simpleExecutor,
		scpiCompleter,
		prompt.OptionTitle("SCPI CLI (SCliPI)"),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(":"))
	p.Run()
}

func ConnectToInstrument(address string) {
	bar := progressbar.New(1000)
	for i := 0; i < 1000; i++ {
		bar.Add(1)
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println("You selected " + address)
}
