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
        MultiBulkReply{
            BulkReply([]byte("SET")), 
            BulkReply([]byte("foo")), 
            BulkReply([]byte("bar")),
        },
        []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
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

var inequalities = []inequality {
    {
        BulkReply([]byte("foo")),
        BulkReply([]byte("bar")),
    },
}

func TestInequality(t *testing.T) {
    for _, ineq := range inequalities {
        if Equal(ineq.lhs, ineq.rhs) {
            t.Fatalf("Equal(%#v, %#v) returned true when false was expected", ineq.lhs, ineq.rhs)
        }
    }
}

