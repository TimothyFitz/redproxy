package redproxy

import (
    "fmt"
    "io"
    "strconv"
    "bytes"
)

type MultiBulkReply []BulkReply
type BulkReply []byte
type SingleLine []byte
type ErrorMessage []byte
type Integer int64

var (
    charMultiBulkReply = []byte("*")
    charBulkReply = []byte("$")
    crlf = []byte("\r\n")
)

func itob(i int) []byte {
    return []byte(strconv.Itoa(i))
}

func panic_type(v interface{}) {
    panic(fmt.Sprintf("Unknown type (%#v)", v))
}

func Equal(i_lhs interface{}, i_rhs interface{}) bool {
    switch lhs := i_lhs.(type) {
    case MultiBulkReply:
        rhs, ok := i_rhs.(MultiBulkReply)
        if !ok { panic_type(i_rhs) }
        for i := range lhs {
            if !Equal(lhs[i], rhs[i]) {
                return false
            }
        }
        return true
    case BulkReply:
        rhs, ok := i_rhs.(BulkReply)
        if !ok { panic_type(i_rhs) }
        return bytes.Equal(lhs, rhs)
    default:
        panic_type(i_lhs)
    }
    return false
}

func Write(iv interface{}, out io.Writer) {
    switch v := iv.(type) {
    case MultiBulkReply:
        out.Write(charMultiBulkReply)
        out.Write(itob(len(v)))
        out.Write(crlf)
        for _, br := range v {
            Write(br, out)
        }
    case BulkReply:
        out.Write(charBulkReply)
        out.Write(itob(len(v)))
        out.Write(crlf)
        out.Write(v)
        out.Write(crlf)
    }
}
