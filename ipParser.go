package main

import (
	"net"
	"strings"
)

type ipNode struct {
	Content  string
	Children []ipNode
}

func parseIpv4(ips []net.IP) ipNode {
	head := ipNode{}
	for _, address := range ips {
		parts := strings.Split(address.String(), ".")
		if exists, node0 := ipNodeExists(head, parts[0]); exists {
			if exists, node1 := ipNodeExists(head.Children[node0], parts[1]); exists {
				if exists, _ := ipNodeExists(head.Children[node0].Children[node1], parts[2]); exists {
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

func ipNodeExists(head ipNode, s string) (bool, int) {
	for i, node := range head.Children {
		if node.Content == s {
			return true, i
		}
	}
	return false, -1
}
