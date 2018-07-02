package ipParser

import (
	"net"
	"strings"
)

type IpParser struct {
}

type IpNode struct {
	Content  string
	Children []IpNode
}

// Only parses the first 2 sections for now
//TODO: Refactor
func ParseIpv4(ips []net.IP) IpNode {
	head := IpNode{}
	for _, address := range ips {
		parts := strings.Split(address.String(), ".")
		if exists, node0 := NodeExists(head, parts[0]); exists {
			if exists, _ := NodeExists(head.Children[node0], parts[1]); exists {
				//Do Nothing
			} else {
				head.Children[node0].Children = append(head.Children[node0].Children, IpNode{Content: parts[1]})
			}
		} else {
			head.Children = append(head.Children, IpNode{Content: parts[0]})
			if len(parts) > 1 {
				head.Children[len(head.Children) - 1].Children = append(head.Children[len(head.Children) - 1].Children, IpNode{Content: parts[1]})
			}
		}
	}
	return head
}

func NodeExists(head IpNode, s string) (bool, int) {
	for i, node := range head.Children {
		if node.Content == s {
			return true, i
		}
	}
	return false, -1
}
