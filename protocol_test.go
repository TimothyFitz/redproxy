package redproxy

import (
    "bytes"
    "testing"
)

type expected_value struct {
    decoded interface{}
    encoded []byte
}

var expected_values = []expected_value{
    {
        MultiBulkData{
            BulkData([]byte("SET")),
            BulkData([]byte("foo")),
            BulkData([]byte("barbar")),
        },
        []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$6\r\nbarbar\r\n"),
    },
    {
        SingleLine([]byte("OK")),
        []byte("+OK\r\n"),
    },
    {
        nil,
        []byte("$-1\r\n"),
    },
    {
        ErrorMessage([]byte("ERR badness")),
        []byte("-ERR badness\r\n"),
    },
}

func encode(iv interface{}) []byte {
    var buff bytes.Buffer
    Write(iv, &buff)
    return buff.Bytes()
}

func TestEncodingKnownGoodValues(t *testing.T) {
    for _, ev := range expected_values {
        output := encode(ev.decoded)
        if !bytes.Equal(ev.encoded, output) {
            t.Fatalf("Expected %#v, got %#v", string(ev.encoded), string(output))
        }
    }
}

func TestEqualityReflexivity(t *testing.T) {
    for _, ev := range expected_values {
        if !Equal(ev.decoded, ev.decoded) {
            t.Fatalf("Equality fail reflexivity test for %#v", ev.decoded)
        }
    }
}

type inequality struct {
    lhs interface{}
    rhs interface{}
}

var inequalities = []inequality{
    {
        BulkData([]byte("foo")),
        BulkData([]byte("bar")),
    },
}

func TestInequality(t *testing.T) {
    for _, ineq := range inequalities {
        if Equal(ineq.lhs, ineq.rhs) {
            t.Fatalf("Equal(%#v, %#v) returned true when false was expected", ineq.lhs, ineq.rhs)
        }
    }
}

func TestDecodingKnownGoodValues(t *testing.T) {
    for _, ev := range expected_values {
        value, err := Read(bytes.NewBuffer(ev.encoded))
        if err != nil {
            t.Fatalf("Unexpected error %#v", err)
        }
        if !Equal(value, ev.decoded) {
            t.Fatalf("Expected %#v, got %#v", ev.decoded, value)
        }
    }
}
