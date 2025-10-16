package main

import (
	"bufio"
	"fmt"
	"github.com/schollz/progressbar"
	"net"
	"strconv"
	"strings"
	"time"
)

type instrument interface {
	connect(string, *progress) error
	reconnect() error
	command(string) error
	query(string) (string, error)
	getSupportedCommands() ([]string, []string, error)
	setTimeout(time.Duration)
	close() error
}

type scpiInstrument struct {
	address    string
	connection *net.TCPConn
	timeout time.Duration
}

func (i *scpiInstrument) connect(address string, p *progress) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	p.forward(20)

	d := net.Dialer{Timeout: i.timeout}

	conn, err := d.Dial("tcp", tcpAddr.String())
	if err != nil {
		return err
	}
	p.forward(20)

	i.address = address
	i.connection = conn.(*net.TCPConn)
	return nil
}

func (i *scpiInstrument) reconnect() error {
	_ = i.close()
	p := &progress{}
	err := i.connect(i.address, p)
	if err != nil {
		return err
	}
	p.forward(60)
	fmt.Println()
	return nil
}

func (i *scpiInstrument) command(command string) error {
	if err := i.exec(command); err != nil {
		return fmt.Errorf("failed to execute the command '%s': %s", command, err)
	}
	return i.queryError()
}

func (i *scpiInstrument) exec(cmd string) error {
	b := []byte(cmd + "\n")
	_ = i.connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := i.connection.Write(b); err != nil {
		return err
	}
	return nil
}

func (i *scpiInstrument) queryError() error {
	res, err := i.query("SYST:ERR?")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(res, "+0") {
		fmt.Println("Error: " + strings.TrimRight(res, "\n"))
		return i.queryError()
	}
	return nil
}

func (i *scpiInstrument) query(cmd string) (res string, err error) {
	if err := i.exec(cmd); err != nil {
		return "", err
	}

	queryCompleted := make(chan bool, 1)
	queryFailed := make(chan bool, 1)
	done := make(chan bool)
	go queryProgress(queryCompleted, queryFailed, done, i.timeout)

	_ = i.connection.SetReadDeadline(time.Now().Add(i.timeout))

	b := bufio.NewReader(i.connection)
	buf := make([]byte, 4096)

	l, err := b.Read(buf)
	if err != nil {
		queryFailed <- true
		<-done
		return "", err
	}

	firstRead := string(buf[:l])

	result := firstRead
	if strings.HasPrefix(firstRead, "#") {
		firstLineIndex := strings.IndexByte(firstRead, '\n')
		result = result[firstLineIndex:]
		responseSize, err := i.parseBlockInfo(firstRead[:firstLineIndex])
		if err != nil {
			queryFailed <- true
			<-done
			return "", err
		}

		numBytesRead := l
		for {
			n, err := b.Read(buf)
			if err != nil {
				queryFailed <- true
				<-done
				return "", err
			}

			numBytesRead += n
			result += string(buf[:n])

			if numBytesRead >= responseSize {
				break
			}
		}
	}
	queryCompleted <- true
	<-done
	return result, nil
}

func (i *scpiInstrument) parseBlockInfo(blockInfo string) (int, error) {
	blockInfo = strings.TrimSuffix(blockInfo, "\n")
	if !strings.HasPrefix(blockInfo, "#") || len(blockInfo) == 0 {
		return -1, fmt.Errorf("unrecognized response header found: %s", blockInfo)
	}

	blockInfo = strings.TrimPrefix(blockInfo, "#")
	characters := strings.Split(blockInfo, "")

	numChars, err := strconv.Atoi(characters[0])
	if err != nil {
		return -1, err
	}

	if len(characters[1:]) != numChars {
		return -1, fmt.Errorf("SCPI block info malformed: %s", blockInfo)
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

func (i *scpiInstrument) setTimeout(timeout time.Duration) {
	i.timeout = timeout
}

func queryProgress(queryCompleted chan bool, queryFailed chan bool, done chan bool, timeout time.Duration) {
	select {
	case <-queryCompleted:
		done <- true
		return
	case <-queryFailed:
		done <- true
		return
	case <-time.After(time.Second):
		break
	}

	bar := progressbar.New(100)
	_ = bar.Add(10)
	percent := 10

loop:
	for {
		select {
		case <-queryCompleted:
			_ = bar.Add(100 - percent)
			break loop
		case <-queryFailed:
			break loop
		case <-time.After(timeout / 10):
			if percent < 90 {
				_ = bar.Add(10)
				percent += 10
			}
		}
	}
	// Clear the progress bar line by moving cursor to start and clearing the line
	fmt.Print("\r\033[K")
	done <- true
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
	timeout time.Duration
}

func (i *simInstrument) connect(address string, p *progress) error {
	p.forward(40)
	return nil
}

func (i *simInstrument) reconnect() error{
	return nil
}

func (i *simInstrument) command(command string) error {
	return nil
}

func (i *simInstrument) query(query string) (string, error) {
	if query == "*ESR?" || query == "*ID?" {
		queryCompleted := make(chan bool, 1)
		queryFailed := make(chan bool, 1)
		done := make(chan bool)
		go queryProgress(queryCompleted, queryFailed, done, i.timeout)
		if query == "*ID?" {
			queryFailed <- true
		} else {
			queryCompleted <- true
		}
		<-done
	}
	return query + "\n", nil
}

func (i *simInstrument) getSupportedCommands() ([]string, []string, error) {
	commands, err := readLinesFromPath("SCPI.txt")
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

func (i *simInstrument) setTimeout(timeout time.Duration) {
	i.timeout = timeout
}

func (i *simInstrument) close() error {
	return nil
}
