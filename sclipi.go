package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"strings"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	elements := strings.Split(d.TextBeforeCursor(), ":")
	var s []prompt.Suggest

	if contains(elements, "FREQuency") {
		s = []prompt.Suggest{
			{Text: "CENTer", Description: "Scpi Command Example"},
		}
	} else {
		s = []prompt.Suggest{
			{Text: "FREQuency", Description: "Scpi Command Example"},
			{Text: "FREQ", Description: "Scpi Command Example"},
			{Text: "articles", Description: "Store the article text posted by user"},
			{Text: "comments", Description: "Store the text commented to articles"},
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursorUntilSeparator(":"), true)
}

func nullCompleter(prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func simpleExecutor(a string) {
	fmt.Println("you entered: " + a)
}

func main() {
	defer fmt.Println("Bye!")
	address := prompt.Input("IP Address: ", nullCompleter)
	fmt.Println("You selected " + address)

	p := prompt.New(simpleExecutor, completer, prompt.OptionCompletionWordSeparator(":"))
	p.Run()
}
