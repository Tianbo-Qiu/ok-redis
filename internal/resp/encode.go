package resp

import "fmt"

func SimpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

func Error(msg string) string {
	return fmt.Sprintf("-%s\r\n", msg)
}

func Integer(n int64) string {
	return fmt.Sprintf(":%d\r\n", n)
}

func BulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

// NilBulk is the RESP null bulk string
// used for a missing value
const NilBulk = "$-1\r\n"
