package main

import (
	"net"
	"bufio"
	"fmt"
	"strings"
)

type iInstrument interface {
	Connect(string) error
	Command(string)
	Query(string) (string, error)
	getSupportedCommands() ([]string, error)
}

type instrument struct {
	address    string
	connection net.Conn
}

func (i instrument) Connect(address string) error {
	i.address = address
	connection, err := net.Dial("tcp", i.address)
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to instrument at address " + i.address)
	i.connection = connection
	return nil
}

func (i instrument) Command(command string) {
	fmt.Fprintf(i.connection, command+"\n")
}

func (i instrument) Query(query string) (string, error) {
	fmt.Fprintf(i.connection, query+"\n")
	message, err := bufio.NewReader(i.connection).ReadString('\n')
	if err != nil {
		return "", err
	}
	return message, nil
}

func (i instrument) getSupportedCommands()  ([]string, error){
	result, err := i.Query("SYST:HELP:HEAD?")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(result, "\n"), nil
}


type simInstrument struct{
	
}

func (i simInstrument) Connect(address string) error {
	return nil
}

func (i simInstrument) Command(command string) {
}

func (i simInstrument) Query(query string) (string, error) {
	return query, nil
}

func (i simInstrument) getSupportedCommands()  ([]string, error){
	lines, err := readLines("MXGSCPI.txt")
	if err != nil {
		return []string{}, err
	}
	return lines, nil
}