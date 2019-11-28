package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type instrument interface {
	connect(time.Duration, string) error
	command(string) error
	query(string) (string, error)
	getSupportedCommands() ([]string, []string, error)
	close() error
}

type scpiInstrument struct {
	address    string
	connection *net.TCPConn
}

func (i *scpiInstrument) connect(timeout time.Duration, address string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	d := net.Dialer{Timeout: timeout}

	conn, err := d.Dial("tcp", tcpAddr.String())
	if err != nil {
		return err
	}

	i.address = address
	i.connection = conn.(*net.TCPConn)
	return nil
}

func (i *scpiInstrument) command(command string) error {
	if err := i.exec(command); err != nil {
		return fmt.Errorf("failed to execute the command '%s': %s", command, err)
	}
	return i.queryError(command)
}

func (i *scpiInstrument) exec(cmd string) error {
	b := []byte(cmd + "\n")
	_ = i.connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := i.connection.Write(b); err != nil {
		return err
	}
	return nil
}

func (i *scpiInstrument) queryError(prevCmd string) error {
	res, err := i.query("SYST:ERR?")
	if err != nil {
		return err
	}
	fmt.Println("Error: " + strings.TrimRight(res, "\n"))
	return nil
}

func (i *scpiInstrument) query(cmd string) (res string, err error) {
	if err := i.exec(cmd); err != nil {
		return "", err
	}

	_ = i.connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	b := bufio.NewReader(i.connection)
	buf := make([]byte, 4096)

	l, err := b.Read(buf)
	if err != nil {
		return "", err
	}

	firstRead := string(buf[:l])

	result := firstRead
	if strings.HasPrefix(firstRead, "#") {
		firstLineIndex := strings.IndexByte(firstRead, '\n')
		result = result[firstLineIndex:]
		responseSize, err := i.parseBlockInfo(firstRead[:firstLineIndex])
		if err != nil {
			return "", err
		}

		numBytesRead := l
		for {
			n, err := b.Read(buf)
			if err != nil {
				return "", err
			}

			numBytesRead += n
			result += string(buf[:n])

			if numBytesRead >= responseSize {
				break
			}
		}
	}
	return result, nil
}

func (i *scpiInstrument) parseBlockInfo(blockInfo string) (int, error) {
	blockInfo = strings.TrimSuffix(blockInfo, "\n")
	if !strings.HasPrefix(blockInfo, "#") || len(blockInfo) == 0 {
		return -1, fmt.Errorf("Unrecognized response header found: " + blockInfo)
	}

	blockInfo = strings.TrimPrefix(blockInfo, "#")
	characters := strings.Split(blockInfo, "")

	numChars, err := strconv.Atoi(characters[0])
	if err != nil {
		return -1, err
	}

	if len(characters[1:]) != numChars {
		return -1, fmt.Errorf("SCPI block info malformed: " + blockInfo)
	}

	resultString := ""
	for _, char := range characters[1 : numChars+1] {
		resultString += char
	}

	result, err := strconv.Atoi(resultString)
	if err != nil {
		return -1, err
	}

	return result, nil
}

func (i *scpiInstrument) getSupportedCommands() ([]string, []string, error) {
	r, err := i.query(":SYST:HELP:HEAD?")
	commands := strings.Split(r, "\n")

	var colonCommands []string
	var starCommands []string
	for _, command := range commands {
		if command != "" {
			if strings.HasPrefix(command, "*") {
				starCommands = append(starCommands, command)
			} else {
				colonCommands = append(colonCommands, command)
			}
		}
	}

	return colonCommands, starCommands, err
}

func (i *scpiInstrument) close() error {
	return i.connection.Close()
}

type simInstrument struct {
}

func (i *simInstrument) connect(timeout time.Duration, address string) error {
	return nil
}

func (i *simInstrument) command(command string) error {
	return nil
}

func (i *simInstrument) query(query string) (string, error) {
	return query + "\n", nil
}

func (i *simInstrument) getSupportedCommands() ([]string, []string, error) {
	commands, err := readLines("SCPI.txt")
	if err != nil {
		return []string{}, []string{}, err
	}

	var colonCommands []string
	var starCommands []string
	for _, command := range commands {
		if command != "" {
			if strings.HasPrefix(command, "*") {
				starCommands = append(starCommands, command)
			} else {
				colonCommands = append(colonCommands, command)
			}
		}
	}
	return colonCommands, starCommands, nil
}

func (i *simInstrument) close() error {
	return nil
}
