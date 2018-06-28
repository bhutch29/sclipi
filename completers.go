package main

import (
	"github.com/c-bata/go-prompt"
	"strings"
	"fmt"
	"net"
	"SCliPI/ipParser"
)

func scpiCompleter(d prompt.Document) []prompt.Suggest {
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

func ipCompleter(d prompt.Document) []prompt.Suggest {
	filtered := filterIps(GetIpAddresses())
	tree := ipParser.ParseIpv4(filtered)

	elements := strings.Split(d.TextBeforeCursor(), ".")
	if contains() { //TODO: Now we have to turn the tree into something useful to generate Suggestions out of

	}
	var s []prompt.Suggest
	return s
}

func nullCompleter(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func filterIps(ips []net.IP) []net.IP {
	var filtered []net.IP
	for _, ip := range ips {
		fmt.Println(ip)
		if !strings.Contains(ip.String(), ":") {
			filtered = append(filtered, ip)
		}
	}
	return filtered
}

func GetIpAddresses() []net.IP {
	var ips [] net.IP
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		addresses, _ := i.Addrs()
		for _, addr := range addresses {
			switch v := addr.(type) {
			case *net.IPNet:
				ips = append(ips, v.IP)
			case *net.IPAddr:
				ips = append(ips, v.IP)
			}
		}
	}
	return ips
}
