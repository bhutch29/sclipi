package main

import (
	"fmt"
)

type ScpiProvider struct {
	tree scpiNode
}

func (p *ScpiProvider) getCommands(i instrument) []string {
	lines, err := i.getSupportedCommands()
	if err != nil {
		fmt.Println(err)
	}
	return lines
}

func (p *ScpiProvider) getTree(i instrument) scpiNode {
	if len(p.tree.Children) == 0 {
		p.tree = parseScpi(p.getCommands(i))
	}
	return p.tree
}
