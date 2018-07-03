package main

import (
	"fmt"
)

type ScpiProvider struct {
	tree scpiNode
}

func (p *ScpiProvider) getCommands() []string {
	lines, err := readLines("MXGSCPI.txt")
	if err != nil {
		fmt.Println(err)
	}
	return lines
}

func (p *ScpiProvider) getTree() scpiNode {
	if len(p.tree.Children) == 0{
		p.tree = parseScpi(p.getCommands())
	}
	return p.tree
}
