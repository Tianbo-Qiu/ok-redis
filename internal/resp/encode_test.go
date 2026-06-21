package resp

import "testing"

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"simple string", SimpleString("OK"), "+OK\r\n"},
		{"error", Error("ERR bad"), "-ERR bad\r\n"},
		{"integer", Integer(42), ":42\r\n"},
		{"bulk string", BulkString("alice"), "$5\r\nalice\r\n"},
		{"empty bulk", BulkString(""), "$0\r\n\r\n"},
		{"nil bulk", NilBulk, "$-1\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
