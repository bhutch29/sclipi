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
	var colors = []string{"DefaultColor", "Black", "DarkRed", "DarkGreen", "Brown", "DarkBlue", "Purple", "Cyan", "LightGray", "DarkGray", "Red", "Green", "Yellow", "Blue", "Fuchsia", "Turquoise", "White",}

	parser := argparse.NewParser("SCliPI", "A SCPI CLI. Features autocomplete and much more")
	ipFlag := parser.String("i", "ip", &argparse.Options{
		Help: "The IP address of the instrument. If not provided, SCliPI will use your network information and auto-completion to assist you"})
	portFlag := parser.String("p", "port", &argparse.Options{
		Help: "The SCPI port of the instrument. Defaults to 5025"})
	scriptFileFlag := parser.String("f", "file", &argparse.Options{
		Help: "The path to a newline-delimited list of commands to be run non-interactively. Must set IP address if using this feature"})
	textColorFlag := parser.Selector("c", "text-color", colors, &argparse.Options{
		Help: "The command line text color"})
	previewColorFlag := parser.Selector("", "preview-color", colors, &argparse.Options{
		Help: "The preview text color"})
	suggestionColorFlag := parser.Selector("", "suggestion-color", colors, &argparse.Options{
		Help: "The suggestion text color"})
	suggestionBgColorFlag := parser.Selector("", "suggestion-bg-color", colors, &argparse.Options{
		Help: "The suggestion bg color"})
	selectedColorFlag := parser.Selector("", "selected-color", colors, &argparse.Options{
		Help: "The selected text color"})
	selectedBgColorFlag := parser.Selector("", "selected-bg-color", colors, &argparse.Options{
		Help: "The selected bg color"})

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

	textColor := parseColor(*textColorFlag, prompt.Yellow)
	suggestionColor := parseColor(*suggestionColorFlag, prompt.White)
	suggestionBgColor := parseColor(*suggestionBgColorFlag, prompt.Cyan)
	selectedColor := parseColor(*selectedColorFlag, prompt.Black)
	selectedBgColor := parseColor(*selectedBgColorFlag, prompt.Turquoise)
	previewColor := parseColor(*previewColorFlag, prompt.Blue)
	p := prompt.New(
		sm.executor,
		sm.completer,
		prompt.OptionTitle("SCPI CLI (SCliPI)"),
		prompt.OptionCompletionWordSeparator(":"),
		prompt.OptionInputTextColor(textColor),
		prompt.OptionSuggestionTextColor(suggestionColor),
		prompt.OptionSuggestionBGColor(suggestionBgColor),
		prompt.OptionSelectedSuggestionTextColor(selectedColor),
		prompt.OptionSelectedSuggestionBGColor(selectedBgColor),
		prompt.OptionDescriptionTextColor(selectedColor),
		prompt.OptionDescriptionBGColor(selectedBgColor),
		prompt.OptionSelectedDescriptionTextColor(suggestionColor),
		prompt.OptionSelectedDescriptionBGColor(suggestionBgColor),
		prompt.OptionPreviewSuggestionTextColor(previewColor))

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

func parseColor(color string, def prompt.Color) prompt.Color {
	if color == ""{
		return def
	} else {
		return colorFromString(color)
	}
}

func colorFromString(color string) prompt.Color {
	switch color{
		case "DefaultColor": return prompt.DefaultColor
		case "Black": return prompt.Black
		case "DarkRed": return prompt.DarkRed
		case "DarkGreen": return prompt.DarkGreen
		case "Brown": return prompt.Brown
		case "DarkBlue": return prompt.DarkBlue
		case "Purple": return prompt.Purple
		case "Cyan": return prompt.Cyan
		case "LightGray": return prompt.LightGray
		case "DarkGray": return prompt.DarkGray
		case "Red": return prompt.Red
		case "Green": return prompt.Green
		case "Yellow": return prompt.Yellow
		case "Blue": return prompt.Blue
		case "Fuchsia": return prompt.Fuchsia
		case "Turquoise": return prompt.Turquoise
		case "White": return prompt.White
		default: log.Fatal("Color not found: " + color)
	}
	return prompt.DefaultColor
}
