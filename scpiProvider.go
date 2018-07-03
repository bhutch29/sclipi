package main

import (
	"fmt"
)

type ScpiProvider struct {
	tree scpiNode
}

func (p *ScpiProvider) getCommands(i iInstrument) []string {
	lines, err := i.getSupportedCommands()
	if err != nil {
		fmt.Println(err)
	}
	return lines
}

func (p *ScpiProvider) getTree(i iInstrument) scpiNode {
	if len(p.tree.Children) == 0{
		p.tree = parseScpi(p.getCommands(i))
	}
	return p.tree
}
