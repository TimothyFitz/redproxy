package redproxy

import (
    "bytes"
    "io"
    "strconv"
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

func Encode(iv interface{}) []byte {
    var buff bytes.Buffer
    Write(iv, &buff)
    return buff.Bytes()
}