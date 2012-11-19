package redproxy

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strconv"
)

type MultiBulkReply []BulkReply
type BulkReply []byte
type SingleLine []byte
type ErrorMessage []byte
type Integer int64

const (
    charMultiBulkReply = '*'
    charBulkReply      = '$'
)

var crlf = []byte("\r\n")

func itob(i int) []byte {
    return []byte(strconv.Itoa(i))
}

func panic_type(v interface{}) {
    panic(fmt.Sprintf("Unknown type (%#v)", v))
}

func Equal(i_lhs interface{}, i_rhs interface{}) bool {
    if i_lhs == nil || i_rhs == nil {
        return false
    }

    switch lhs := i_lhs.(type) {
    case MultiBulkReply:
        rhs, ok := i_rhs.(MultiBulkReply)
        if !ok {
            panic_type(i_rhs)
        }
        for i := range lhs {
            if !Equal(lhs[i], rhs[i]) {
                return false
            }
        }
        return true
    case BulkReply:
        rhs, ok := i_rhs.(BulkReply)
        if !ok {
            panic_type(i_rhs)
        }
        return bytes.Equal(lhs, rhs)
    default:
        panic_type(i_lhs)
    }
    return false
}

func Write(iv interface{}, out io.Writer) {
    switch v := iv.(type) {
    case MultiBulkReply:
        out.Write([]byte{charMultiBulkReply})
        out.Write(itob(len(v)))
        out.Write(crlf)
        for _, br := range v {
            Write(br, out)
        }
    case BulkReply:
        out.Write([]byte{charBulkReply})
        out.Write(itob(len(v)))
        out.Write(crlf)
        out.Write(v)
        out.Write(crlf)
    }
}

type ProtocolError struct {
    message string
}

func (pe *ProtocolError) Error() string {
    return pe.message
}

func newProtocolError(format string, v ...interface{}) (pe *ProtocolError) {
    pe = new(ProtocolError)
    pe.message = fmt.Sprintf(format, v...)
    return
}

func readOrError(in *bufio.Reader, slice []byte) error {
    pos := 0
    for pos < len(slice) {
        count, err := in.Read(slice[pos:len(slice)])
        pos += count
        if err != nil {
            return err
        }
    }
    return nil
}

func read(in *bufio.Reader) (interface{}, error) {
    header, err := in.ReadBytes('\n')

    if err != nil {
        return nil, err
    }

    if header[len(header)-1] != '\r' {
        err = newProtocolError(fmt.Sprintf("Invalid reply: %#v", header))
    }

    msg_type := header[0]
    msg_body := header[1 : len(header)-2]

    switch msg_type {
    case charBulkReply:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }

        data := make([]byte, length)
        err = readOrError(in, data)
        if err != nil {
            return nil, err
        }
        expected_crlf := make([]byte, 2)
        err = readOrError(in, expected_crlf)
        if err != nil {
            return nil, err
        }

        if !bytes.Equal(expected_crlf, crlf) {
            return nil, newProtocolError(fmt.Sprintf("Expected crlf, got: %#v", expected_crlf))
        }

        return BulkReply(data), nil
    case charMultiBulkReply:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }
        mbr := make(MultiBulkReply, length)
        for i := 0; i < length; i++ {
            v, err := read(in)
            if err != nil {
                return nil, err
            }
            br, ok := v.(BulkReply)
            if !ok {
                err = newProtocolError(fmt.Sprintf("Unexpected non-bulk reply: %#v", v))
            }
            mbr[i] = br
        }
        return mbr, nil
    }
    return nil, newProtocolError(fmt.Sprintf("Unknown reply type: %#v", header))
}

func Read(in io.Reader) (v interface{}, err error) {
    v, err = read(bufio.NewReader(in))
    return
}
