package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type instrument interface {
	Connect(string) error
	Command(string)
	Query(string) (string, error)
	getSupportedCommands() ([]string, error)
}

type scpiInstrument struct {
	address    string
	connection net.Conn
}

func (i scpiInstrument) Connect(address string) error {
	i.address = address
	connection, err := net.Dial("tcp", i.address)
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to scpiInstrument at address " + i.address)
	i.connection = connection
	return nil
}

func (i scpiInstrument) Command(command string) {
	fmt.Fprintf(i.connection, command+"\n")
}

func (i scpiInstrument) Query(query string) (string, error) {
	fmt.Fprintf(i.connection, query+"\n")
	message, err := bufio.NewReader(i.connection).ReadString('\n')
	if err != nil {
		return "", err
	}
	return message, nil
}

func (i scpiInstrument) getSupportedCommands() ([]string, error) {
	result, err := i.Query("SYST:HELP:HEAD?")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(result, "\n"), nil
}

type simInstrument struct {
}

func (i simInstrument) Connect(address string) error {
	return nil
}

func (i simInstrument) Command(command string) {
}

func (i simInstrument) Query(query string) (string, error) {
	return query, nil
}

func (i simInstrument) getSupportedCommands() ([]string, error) {
	lines, err := readLines("SCPI.txt")
	if err != nil {
		return []string{}, err
	}
	return lines, nil
}
