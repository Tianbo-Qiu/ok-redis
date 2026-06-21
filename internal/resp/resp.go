// Package resp implements parsing and encoding of
// the Redis serialization protocol (RESP)
package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func readBulkString(r *bufio.Reader) (string, error) {
	header, err := readLine(r)
	if err != nil {
		return "", err
	}

	if len(header) == 0 || header[0] != '$' {
		return "", fmt.Errorf("expected '$', got %q", header)
	}

	n, err := strconv.Atoi(header[1:])
	if err != nil {
		return "", fmt.Errorf("invalid bulk length %q: %w", header[1:], err)
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}

	// consume the trailing \r\n
	// we just need to consume those two bytes so the next call starts cleanly
	if _, err := readLine(r); err != nil {
		return "", err
	}

	return string(buf), nil
}

// ReadCommand reads one full client command.
// Redis clients send commands as a RESP array of bulk strings,
// e.g. "SET name alice" arrives on the wire as:
// *3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nalice\r\n
//
// ReadCommand returns the parts as a slice: []string{"SET", "name", "alice"}
func ReadCommand(r *bufio.Reader) ([]string, error) {
	header, err := readLine(r)
	if err != nil {
		return nil, err
	}

	if len(header) == 0 || header[0] != '*' {
		return nil, fmt.Errorf("expected '*', got %q", header)
	}

	count, err := strconv.Atoi(header[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length %q: %w", header[1:], err)
	}

	args := make([]string, 0, count)
	for range count {
		s, err := readBulkString(r)
		if err != nil {
			return nil, err
		}
		args = append(args, s)
	}

	return args, nil
}
