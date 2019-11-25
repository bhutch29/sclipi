package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type instrument interface {
	Connect(time.Duration, string) error
	Command(string) error
	Query(string) (string, error)
	getSupportedCommands() ([]string, error)
	Close() error
}

type scpiInstrument struct {
	address    string
	connection *net.TCPConn
}

func (i *scpiInstrument) Connect(timeout time.Duration, address string) error {
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

func (i *scpiInstrument) Command(command string) error {
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
	res, err := i.Query("SYST:ERR?")
	if err != nil {
		return err
	}
	fmt.Println("Error: " + strings.TrimRight(res, "\n"))
	return nil
}

func (i *scpiInstrument) Query(cmd string) (res string, err error) {
	if err := i.exec(cmd); err != nil {
		return "", err
	}

	buf := make([]byte, 4096)
	_ = i.connection.SetReadDeadline(time.Now().Add(10 * time.Second))
	l, err := i.connection.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:l]), nil
}

func (i *scpiInstrument) getSupportedCommands() ([]string, error) {
	if err := i.exec(":SYST:HELP:HEAD?"); err != nil {
		return []string{}, err
	}

	var buf []byte
	for {
		temp := make([]byte, 4096)
		l, err := i.connection.Read(temp)
		if err != nil {
			return []string{}, err
		}
		buf = append(buf, temp[:l]...)
		if strings.Contains(string(buf), "*WAI") {
			break
		}
	}

	commands := strings.Split(string(buf), "\n")
	var result []string
	for _, command := range commands {
		if command != "" && command[0] != '#' {
			result = append(result, command)
		}
	}
	return result, nil
}

//TODO: Utilize this
// func (i *scpiInstrument) bulkExec(cmds ...string) error {
// 	cmd := strings.Join(cmds, ";")
// 	return i.exec(cmd)
// }

func (i *scpiInstrument) Close() error {
	return i.connection.Close()
}

type simInstrument struct {
}

func (i *simInstrument) Connect(timeout time.Duration, address string) error {
	return nil
}

func (i *simInstrument) Command(command string) error {
	return nil
}

func (i *simInstrument) Query(query string) (string, error) {
	return query + "\n", nil
}

func (i *simInstrument) getSupportedCommands() ([]string, error) {
	lines, err := readLines("SCPI.txt")
	if err != nil {
		return []string{}, err
	}
	return lines, nil
}

func (i *simInstrument) Close() error {
	return nil
}
