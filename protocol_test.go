package redproxy

import (
    "bufio"
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
    {
        Integer(99),
        []byte(":99\r\n"),
    },
}

var old_protocol_expected_values = []expected_value{
    {
        MultiBulkData{
            BulkData([]byte("PING")),
        },
        []byte("PING\r\n"),
    },
    {
        MultiBulkData{
            BulkData([]byte("SET")),
            BulkData([]byte("foo")),
            BulkData([]byte("barbar")),
        },
        []byte("SET foo 6\r\nbarbar\r\n"),
    },
}

func encode(iv interface{}) []byte {
    var buff bytes.Buffer
    output := bufio.NewWriter(&buff)
    Write(iv, output)
    output.Flush()
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

func concat(lhs, rhs []expected_value) []expected_value {
    result := make([]expected_value, len(lhs)+len(rhs))
    copy(result, lhs)
    copy(result[len(lhs):], rhs)
    return result
}

func TestDecodingKnownGoodValues(t *testing.T) {
    for _, ev := range concat(expected_values, old_protocol_expected_values) {
        value, err := Read(bufio.NewReader(bytes.NewBuffer(ev.encoded)))
        if err != nil {
            t.Fatalf("Unexpected error %#v", err)
        }
        if !Equal(value, ev.decoded) {
            t.Fatalf("Expected %#v, got %#v", ev.decoded, value)
        }
    }
}

func BenchmarkDecodingSmallBulkData(b *testing.B) {
    b.StopTimer()
    piece := encode(BulkData([]byte("xxx")))
    data := bytes.Repeat(piece, b.N)
    input := bufio.NewReader(bytes.NewBuffer(data))
    b.StartTimer()
    for i := 0; i < b.N; i++ {
        _, err := Read(input)
        if err != nil {
            b.Fatal("Unexpected error:", err)
        }
    }
}