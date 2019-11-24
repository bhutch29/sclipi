package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/c-bata/go-prompt"
	"log"
	"os"
	"strings"
)

var colors = []string{"DefaultColor", "Black", "DarkRed", "DarkGreen", "Brown", "DarkBlue", "Purple", "Cyan",
	"LightGray", "DarkGray", "Red", "Green", "Yellow", "Blue", "Fuchsia", "Turquoise", "White",}

type Arguments struct{
	Ip *string
	Port *string
	ScriptFile *string
	Silent *bool
	TextColor prompt.Color
	PreviewColor prompt.Color
	SuggestionColor prompt.Color
	SuggestionBgColor prompt.Color
	SelectedColor prompt.Color
	SelectedBgColor prompt.Color
}

func ParseArgs() Arguments {
	args := Arguments{}
	parser := argparse.NewParser("Sclipi", "A SCPI cli. Features autocomplete and much more")
	args.Ip = parser.String("i", "ip", &argparse.Options{
		Help: "The IP address of the instrument. If not provided, Sclipi will use your network information and auto-completion to assist you"})
	args.Port = parser.String("p", "port", &argparse.Options{
		Default: "5025",
		Help: "The SCPI port of the instrument"})
	args.ScriptFile = parser.String("f", "file", &argparse.Options{
		Help: "The path to a newline-delimited list of commands to be run non-interactively. Must set IP address if using this feature"})
	args.Silent = parser.Flag("s", "silent", &argparse.Options{
		Help: "Suppresses unnecessary output"})
	textColorFlag := parser.Selector("c", "text-color", colors, &argparse.Options{
		Default: colors[prompt.Yellow],
		Help: "The command line text color"})
	previewColorFlag := parser.Selector("", "preview-color", colors, &argparse.Options{
		Default: colors[prompt.Blue],
		Help: "The preview text color"})
	suggestionColorFlag := parser.Selector("", "suggestion-color", colors, &argparse.Options{
		Default: colors[prompt.White],
		Help: "The suggestion text color"})
	suggestionBgColorFlag := parser.Selector("", "suggestion-bg-color", colors, &argparse.Options{
		Default: colors[prompt.DarkBlue],
		Help: "The suggestion bg color"})
	selectedColorFlag := parser.Selector("", "selected-color", colors, &argparse.Options{
		Default: colors[prompt.Black],
		Help: "The selected text color"})
	selectedBgColorFlag := parser.Selector("", "selected-bg-color", colors, &argparse.Options{
		Default: colors[prompt.Cyan],
		Help: "The selected bg color"})

	parser.HelpFunc = HelpMessage

	if err := parser.Parse(os.Args); err != nil {
		log.Fatal(parser.Usage(err))
	}

	args.TextColor = colorFromString(*textColorFlag)
	args.SuggestionColor = colorFromString(*suggestionColorFlag)
	args.SuggestionBgColor = colorFromString(*suggestionBgColorFlag)
	args.SelectedColor = colorFromString(*selectedColorFlag)
	args.SelectedBgColor = colorFromString(*selectedBgColorFlag)
	args.PreviewColor = colorFromString(*previewColorFlag)

	return args
}

func HelpMessage(o *argparse.Command, _ interface{}) string {
	var result string

	maxWidth := 80

	arguments := make([]argparse.Arg, 0)
	current := o
	for current != nil {
		if current.GetArgs() != nil {
			arguments = append(arguments, current.GetArgs()...)
		}
		current = current.GetParent()
	}

	result = addToLastLine(result, o.GetDescription(), maxWidth, 0, true)
	result = result + "\n\n"

	var argPadding int
	if len(arguments) > 0 {
		argContent := "Arguments:\n\n"
		// Find biggest padding
		for _, argument := range arguments {
			if len(argument.GetLname())+9 > argPadding {
				argPadding = len(argument.GetLname()) + 9
			}
		}
		// Now add args with padding
		for _, argument := range arguments {
			arg := "  "
			if argument.GetSname() != "" {
				arg += "-" + argument.GetSname() + "  "
			} else {
				arg += "    "
			}
			arg += "--" + argument.GetLname()
			arg += strings.Repeat(" ", argPadding-len(arg))
			if argument.GetOpts() != nil && argument.GetOpts().Help != "" {
				arg = addToLastLine(arg, getHelpMessage(argument), maxWidth, argPadding, true)
			}
			argContent += arg + "\n"
		}
		result += argContent + "\n\n"
	}

	var temp = "Color Options: "
	temp += strings.Repeat(" ", argPadding-len(temp))
	for _, color := range colors {
		temp += "\"" + color +  "\", "
	}
	temp = strings.TrimSuffix(temp, ", ")
	result = addToLastLine(result, temp, maxWidth, argPadding, true)
	return result + "\n"
}

func getHelpMessage(o argparse.Arg) string {
	message := ""
	if len(o.GetOpts().Help) > 0 {
		message += o.GetOpts().Help
		if !o.GetOpts().Required && o.GetOpts().Default != nil {
			message += fmt.Sprintf(". Default: %v", o.GetOpts().Default)
		}
	}
	return message
}

func addToLastLine(base string, add string, width int, padding int, canSplit bool) string {
	// If last line has less than 10% space left, do not try to fill in by splitting else just try to split
	hasTen := (width - len(getLastLine(base))) > width/10
	if len(getLastLine(base)+" "+add) >= width {
		if hasTen && canSplit {
			adds := strings.Split(add, " ")
			for _, v := range adds {
				base = addToLastLine(base, v, width, padding, false)
			}
			return base
		}
		base = base + "\n" + strings.Repeat(" ", padding)
	}
	base = base + " " + add
	return base
}

func getLastLine(input string) string {
	slice := strings.Split(input, "\n")
	return slice[len(slice)-1]
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
