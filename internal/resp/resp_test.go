package resp

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestReadLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("PING\r\nfoo\r\n"))

	got, err := readLine(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "PING" {
		t.Errorf("got %q, want %q", got, "PING")
	}
}

func TestReadBulkString(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("$5\r\nalice\r\n"))

	got, err := readBulkString(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "alice" {
		t.Errorf("got %q, want %q", got, "alice")
	}
}

func TestReadCommand(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nalice\r\n"
	r := bufio.NewReader(strings.NewReader(input))

	got, err := ReadCommand(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"SET", "name", "alice"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
