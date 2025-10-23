package utils

import (
	"bufio"
	"errors"
	"fmt"
  "hash/fnv"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar"
)

var ErrConnectionClosed = errors.New("connection closed")

type Instrument interface {
	Connect(string, func(int)) error
	Command(string) error
	Query(string) (string, error)
	GetSupportedCommandsTree() (ScpiNode, ScpiNode, error)
	SetTimeout(time.Duration)
	QueryError([]string) ([]string, error)
	Close() error
}

type scpiInstrument struct {
	address     string
	connection  *net.TCPConn
	timeout     time.Duration
  mu          sync.RWMutex
  interactive bool
  headersHash uint32
  starTree    ScpiNode
  colonTree   ScpiNode
}

func NewScpiInstrument(timeout time.Duration, interactive bool) Instrument {
  return &scpiInstrument{timeout: timeout, interactive: interactive}
}

func (i *scpiInstrument) Connect(address string, progress func(int)) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	if progress != nil {
		progress(20)
	}

	d := net.Dialer{Timeout: i.timeout}

	conn, err := d.Dial("tcp", tcpAddr.String())
	if err != nil {
		return err
	}
	if progress != nil {
		progress(20)
	}

	i.address = address
	i.connection = conn.(*net.TCPConn)
	return nil
}

func (i *scpiInstrument) Command(command string) error {
	if err := i.exec(command); err != nil {
		return fmt.Errorf("failed to execute the command '%s': %s", command, err)
	}
	return nil
}

func (i *scpiInstrument) exec(cmd string) error {
  i.mu.Lock()
  defer i.mu.Unlock()
	b := []byte(cmd + "\n")
	_ = i.connection.SetWriteDeadline(time.Now().Add(i.timeout))
	if _, err := i.connection.Write(b); err != nil {
		if isConnectionError(err) {
			return fmt.Errorf("%w: %v", ErrConnectionClosed, err)
		}
		return err
	}
	return nil
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "use of closed network connection") ||
		strings.TrimSpace(errStr) == "EOF" ||
		errors.Is(err, net.ErrClosed)
}

func (i *scpiInstrument) QueryError(errors []string) ([]string, error) {
	res, err := i.Query("SYST:ERR?")
	if err != nil {
		return errors, err
	}
	if !strings.HasPrefix(res, "+0") {
		errors = append(errors, strings.TrimRight(res, "\n"))
		return i.QueryError(errors)
	}
	return errors, nil
}

func (i *scpiInstrument) Query(cmd string) (res string, err error) {
	if err := i.exec(cmd); err != nil {
		return "", err
	}

	queryCompleted := make(chan bool, 1)
	queryFailed := make(chan bool, 1)
	done := make(chan bool)
	go queryProgress(queryCompleted, queryFailed, done, i.timeout, i.interactive)

  i.mu.Lock()
  defer i.mu.Unlock()
	_ = i.connection.SetReadDeadline(time.Now().Add(i.timeout))

	b := bufio.NewReader(i.connection)
	buf := make([]byte, 4096)

	l, err := b.Read(buf)
	if err != nil {
		queryFailed <- true
		<-done
		if isConnectionError(err) {
			return "", fmt.Errorf("%w: %v", ErrConnectionClosed, err)
		}
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
				if isConnectionError(err) {
					return "", fmt.Errorf("%w: %v", ErrConnectionClosed, err)
				}
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

func (i *scpiInstrument) SetTimeout(timeout time.Duration) {
	i.timeout = timeout
}

func queryProgress(queryCompleted chan bool, queryFailed chan bool, done chan bool, timeout time.Duration, interactive bool) {
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

  var bar *progressbar.ProgressBar
  var percent int
  if interactive {
	  bar = progressbar.New(100)
	  _ = bar.Add(10)
	  percent = 10
  }

loop:
	for {
		select {
		case <-queryCompleted:
      if interactive {
			  _ = bar.Add(100 - percent)
      }
			break loop
		case <-queryFailed:
			break loop
		case <-time.After(timeout / 10):
			if interactive && percent < 90 {
				_ = bar.Add(10)
				percent += 10
			}
		}
	}
	// Clear the progress bar line by moving cursor to start and clearing the line
	fmt.Print("\r\033[K")
	done <- true
}

func (i *scpiInstrument) getSupportedCommands() ([]string, []string, uint32, error) {
	r, err := i.Query(":SYST:HELP:HEAD?")
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

  hash := hash(r);

	return starCommands, colonCommands, hash, err
}

func (i *scpiInstrument) GetSupportedCommandsTree() (ScpiNode, ScpiNode, error) {
  starCommands, colonCommands, hash, err := i.getSupportedCommands()
  if err != nil {
    return ScpiNode{}, ScpiNode{}, err
  }
  if (hash != i.headersHash) {
    i.starTree = parseScpi(starCommands)
    i.colonTree = parseScpi(colonCommands)
    i.headersHash = hash
  }
  return i.starTree, i.colonTree, nil
}

func (i *scpiInstrument) Close() error {
	return i.connection.Close()
}

type simInstrument struct {
	timeout time.Duration
  interactive bool
}

func NewSimInstrument(timeout time.Duration, interactive bool) Instrument {
  return &simInstrument{timeout: timeout, interactive: interactive}
}

func (i *simInstrument) Connect(address string, progress func(int)) error {
	if progress != nil {
		progress(40)
	}
	return nil
}

func (i *simInstrument) Command(command string) error {
	return nil
}

func (i *simInstrument) Query(query string) (string, error) {
	if query == "*ESR?" || query == "*ID?" {
		queryCompleted := make(chan bool, 1)
		queryFailed := make(chan bool, 1)
		done := make(chan bool)
		go queryProgress(queryCompleted, queryFailed, done, i.timeout, i.interactive)
		if query == "*ID?" {
			queryFailed <- true
		} else {
			queryCompleted <- true
		}
		<-done
	}
	return query + "\n", nil
}

func (i *simInstrument) getSupportedCommands() ([]string, []string, uint32, error) {
	commands, err := readLinesFromPath("SCPI.txt")
	if err != nil {
		return []string{}, []string{}, 0, err
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

  fakeHash := uint32(1234)
	return starCommands, colonCommands, fakeHash, nil
}

func (i *simInstrument) GetSupportedCommandsTree() (ScpiNode, ScpiNode, error) {
  starCommands, colonCommands, _, err := i.getSupportedCommands()
  if err != nil {
    return ScpiNode{}, ScpiNode{}, err
  }
  starTree := parseScpi(starCommands);
  colonTree := parseScpi(colonCommands);
  return starTree, colonTree, nil
}

func (i *simInstrument) SetTimeout(timeout time.Duration) {
	i.timeout = timeout
}

func (i *simInstrument) QueryError(errors []string) ([]string, error) {
	return errors, nil
}

func (i *simInstrument) Close() error {
	return nil
}

func readLinesFromPath(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}
