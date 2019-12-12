package main

import (
	"github.com/c-bata/go-prompt"
	"net"
	"strings"
)

type ipCompleter struct {
	addresses    []net.IP
	simSupported bool
	ipProvider   func(func([]net.IP) []net.IP) []net.IP
}

func newIpCompleter(simSupported bool) ipCompleter {
	return ipCompleter{simSupported: simSupported, ipProvider: getIPv4InterfaceAddresses}
}

func (ip *ipCompleter) completer(d prompt.Document) []prompt.Suggest {
	tree := ip.parseIpv4(ip.getIpAddresses(ip.filterIpv4))
	inputs := strings.Split(d.TextBeforeCursor(), ".")
	s := ip.suggestsFromNode(ip.getCurrentNode(tree, inputs))
	ipSuggests := prompt.FilterHasPrefix(s, d.GetWordBeforeCursorUntilSeparator("."), true)

	var o []prompt.Suggest
	if ip.simSupported {
		o = []prompt.Suggest{{Text: "simulated", Description: "Simulate SCPI instrument using SCPI.txt file in Sclipi directory"}}
	}
	o = append(o, []prompt.Suggest{{Text: "localhost", Description: "Connect to local machine"}, {Text: "?", Description: "Help"}}...)
	otherSuggests := prompt.FilterHasPrefix(o, d.GetWordBeforeCursor(), false)

	return prompt.FilterHasPrefix(append(ipSuggests, otherSuggests...), d.GetWordBeforeCursorUntilSeparator("."), true)
}

func (ip *ipCompleter) getCurrentNode(node ipNode, inputs []string) ipNode {
	current := node
	next := node
	for i, item := range inputs {
		if success, node := ip.getNodeChildByContent(current, item); success { // Found a match, store it away and keep looking in case period has not been pressed
			current = next
			next = node
			continue
		} else if i < len(inputs) -1 { // Period pressed twice in a row, return nothing
			return ipNode{}
		}
		current = next // item not found
		break
	}
	// If we made it here without calling break, then a match has been found but period has not been pressed yet, so return the previous node's suggestions
	return current
}

func (ip *ipCompleter) getNodeChildByContent(parent ipNode, item string) (bool, ipNode) {
	for _, node := range parent.Children {
		if node.Content == item {
			return true, node
		}
	}
	return false, ipNode{}
}

func (ip *ipCompleter) suggestsFromNode(node ipNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children {
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func (ip *ipCompleter) getIpAddresses(filter func([]net.IP) []net.IP) []net.IP {
	if len(ip.addresses) == 0 {
		ip.addresses = ip.ipProvider(filter)
	}
	return filter(ip.addresses)
}

func getIPv4InterfaceAddresses(filter func([]net.IP) []net.IP) []net.IP {
	var ips []net.IP
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		addresses, _ := i.Addrs()
		for _, addr := range addresses {
			if strings.HasPrefix(addr.String(), "127") {
				continue
			}
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

func (ip *ipCompleter) filterIpv4(ips []net.IP) []net.IP {
	var filtered []net.IP
	for _, ip := range ips {
		if !strings.Contains(ip.String(), ":") {
			filtered = append(filtered, ip)
		}
	}
	return filtered
}

type ipNode struct {
	Content  string
	Children []ipNode
}

func (ip *ipCompleter) parseIpv4(ips []net.IP) ipNode {
	head := ipNode{}
	for _, address := range ips {
		parts := strings.Split(address.String(), ".")
		if exists, node0 := ip.ipNodeExists(head, parts[0]); exists {
			if exists, node1 := ip.ipNodeExists(head.Children[node0], parts[1]); exists {
				if exists, _ := ip.ipNodeExists(head.Children[node0].Children[node1], parts[2]); exists {
					//Do Nothing
				} else {
					head.Children[node0].Children[node1].Children = append(head.Children[node0].Children[node1].Children, ipNode{Content: parts[2]})
				}
			} else {
				head.Children[node0].Children = append(head.Children[node0].Children, ipNode{Content: parts[1]})
				last := len(head.Children[node0].Children) - 1
				head.Children[node0].Children[last].Children = append(head.Children[node0].Children[last].Children, ipNode{Content: parts[2]})
			}
		} else {
			head.Children = append(head.Children, ipNode{Content: parts[0]})
			last := len(head.Children) - 1
			head.Children[last].Children = append(head.Children[last].Children, ipNode{Content: parts[1]})
			last2 := len(head.Children[last].Children) - 1
			head.Children[last].Children[last2].Children = append(head.Children[last].Children[last2].Children, ipNode{Content: parts[2]})
		}
	}
	return head
}

func (ip *ipCompleter) ipNodeExists(head ipNode, s string) (bool, int) {
	for i, node := range head.Children {
		if node.Content == s {
			return true, i
		}
	}
	return false, -1
}
