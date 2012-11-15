package redproxy

import (
    "bytes"
    "testing"
)

type expected_value struct {
    input interface{}
    output []byte
}

var expected_values = []expected_value{
    {
        MultiBulkReply{
            BulkReply([]byte("SET")), 
            BulkReply([]byte("foo")), 
            BulkReply([]byte("bar")),
        },
        []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
    },
}

func TestExpectedValues(t *testing.T) {
    for _, ev := range expected_values {
        output := Encode(ev.input)
        if !bytes.Equal(output, ev.output) {
            t.Fatalf("Expected %v, got %v", string(ev.output), string(output))
        }
    }
}