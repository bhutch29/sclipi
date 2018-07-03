package main

import (
	"net"
	"strings"
)

type ipNode struct {
	Content  string
	Children []ipNode
}

// Only parses the first 2 sections for now
//TODO: Refactor
func parseIpv4(ips []net.IP) ipNode {
	head := ipNode{}
	for _, address := range ips {
		parts := strings.Split(address.String(), ".")
		if exists, node0 := ipNodeExists(head, parts[0]); exists {
			if exists, _ := ipNodeExists(head.Children[node0], parts[1]); exists {
				//Do Nothing
			} else {
				head.Children[node0].Children = append(head.Children[node0].Children, ipNode{Content: parts[1]})
			}
		} else {
			head.Children = append(head.Children, ipNode{Content: parts[0]})
			if len(parts) > 1 {
				head.Children[len(head.Children)-1].Children = append(head.Children[len(head.Children)-1].Children, ipNode{Content: parts[1]})
			}
		}
	}
	return head
}

func ipNodeExists(head ipNode, s string) (bool, int) {
	for i, node := range head.Children {
		if node.Content == s {
			return true, i
		}
	}
	return false, -1
}
