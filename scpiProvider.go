package main

import (
	"SCliPI/scpiParser"
	"fmt"
)

type ScpiProvider struct {
	tree scpiParser.ScpiNode
}

func (p *ScpiProvider) getCommands() []string {
	lines, err := scpiParser.ReadLines("scpiParser/MXGSCPI.txt")
	if err != nil {
		fmt.Println(err)
	}
	return lines
}

func (p *ScpiProvider) GetTree() scpiParser.ScpiNode {
	if len(p.tree.Children) == 0{
		p.tree = scpiParser.Parse(p.getCommands())
	}
	return p.tree
}
