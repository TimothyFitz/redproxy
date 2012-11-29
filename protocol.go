package redproxy

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strconv"
    "strings"
)

type MultiBulkData []BulkData
type BulkData []byte
type SingleLine []byte
type ErrorMessage []byte
type Integer int64

func (bd BulkData) String() string {
    return fmt.Sprintf("BulkData(%s)", string(bd))
}

func (mbd MultiBulkData) String() string {
    pieces := make([]string, len(mbd))
    for i, piece := range mbd {
        pieces[i] = piece.String()
    }
    return fmt.Sprintf("MultiBulkData[%s]", strings.Join(pieces, ", "))
}

func (sl SingleLine) String() string {
    return fmt.Sprintf("SingleLine(%s)", string(sl))
}

func (em ErrorMessage) String() string {
    return fmt.Sprintf("ErrorMessage(%s)", string(em))
}

func (i Integer) String() string {
    return fmt.Sprintf("Integer(%d)", int64(i))
}

const (
    charMultyBulkData = '*'
    charBulkData      = '$'
    charSingleLine    = '+'
    charErrorMessage  = '-'
    charInteger       = ':'
)

var bulkCommands = map[string]bool{
    "zrevrank":  true,
    "zrem":      true,
    "echo":      true,
    "config":    true,
    "lset":      true,
    "set":       true,
    "setex":     true,
    "append":    true,
    "hsetnx":    true,
    "publish":   true,
    "sismember": true,
    "lpush":     true,
    "hmget":     true,
    "hdel":      true,
    "hexists":   true,
    "rpush":     true,
    "zscore":    true,
    "setnx":     true,
    "zadd":      true,
    "hset":      true,
    "hget":      true,
    "zincrby":   true,
    "mset":      true,
    "smove":     true,
    "hmset":     true,
    "getset":    true,
    "zrank":     true,
    "sadd":      true,
    "srem":      true,
    "msetnx":    true,
    "lrem":      true,
}

var crlf = []byte("\r\n")

func itob(i int) []byte {
    return []byte(strconv.Itoa(i))
}

func panic_type(v interface{}) {
    panic(fmt.Sprintf("Unknown type (%#v)", v))
}

func Equal(i_lhs interface{}, i_rhs interface{}) bool {
    switch lhs := i_lhs.(type) {
    case MultiBulkData:
        rhs, ok := i_rhs.(MultiBulkData)
        if !ok {
            return false
        }
        for i := range lhs {
            if !Equal(lhs[i], rhs[i]) {
                return false
            }
        }
        return true
    case BulkData:
        rhs, ok := i_rhs.(BulkData)
        if !ok {
            return false
        }
        return bytes.Equal(lhs, rhs)
    case SingleLine:
        rhs, ok := i_rhs.(SingleLine)
        if !ok {
            return false
        }
        return bytes.Equal(lhs, rhs)
    case ErrorMessage:
        rhs, ok := i_rhs.(ErrorMessage)
        if !ok {
            return false
        }
        return bytes.Equal(lhs, rhs)
    case Integer:
        rhs, ok := i_rhs.(Integer)
        if !ok {
            return false
        }
        return lhs == rhs
    case nil:
        return i_rhs == nil
    default:
        panic_type(i_lhs)
    }
    return false
}

func Write(iv interface{}, out io.Writer) {
    switch v := iv.(type) {
    case MultiBulkData:
        out.Write([]byte{charMultyBulkData})
        out.Write(itob(len(v)))
        out.Write(crlf)
        for _, br := range v {
            Write(br, out)
        }
    case BulkData:
        out.Write([]byte{charBulkData})
        out.Write(itob(len(v)))
        out.Write(crlf)
        out.Write(v)
        out.Write(crlf)
    case SingleLine:
        out.Write([]byte{charSingleLine})
        out.Write(v)
        out.Write(crlf)
    case nil:
        out.Write([]byte("$-1\r\n"))
    case ErrorMessage:
        out.Write([]byte{charErrorMessage})
        out.Write(v)
        out.Write(crlf)
    case Integer:
        out.Write([]byte{charInteger})
        out.Write(itob(int(v)))
        out.Write(crlf)
    default:
        panic_type(iv)
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

func expectCrlf(in *bufio.Reader) error {
    expected_crlf := make([]byte, 2)
    err := readOrError(in, expected_crlf)
    if err != nil {
        return err
    }

    if !bytes.Equal(expected_crlf, crlf) {
        return newProtocolError(fmt.Sprintf("Expected crlf, got: %#v", expected_crlf))
    }

    return nil
}

func Read(in *bufio.Reader) (interface{}, error) {
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
    case charBulkData:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }

        if length < 0 {
            return nil, nil
        }

        data := make([]byte, length)
        if err = readOrError(in, data); err != nil {
            return nil, err
        }

        if err = expectCrlf(in); err != nil {
            return nil, err
        }

        return BulkData(data), nil

    case charMultyBulkData:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }
        mbd := make(MultiBulkData, length)
        for i := 0; i < length; i++ {
            v, err := Read(in)
            if err != nil {
                return nil, err
            }
            br, ok := v.(BulkData)
            if !ok {
                err = newProtocolError(fmt.Sprintf("Unexpected non-bulk reply: %#v", v))
            }
            mbd[i] = br
        }
        return mbd, nil

    case charSingleLine:
        sl := make(SingleLine, len(msg_body))
        copy(sl, msg_body)
        return sl, nil

    case charErrorMessage:
        em := make(ErrorMessage, len(msg_body))
        copy(em, msg_body)
        return em, nil

    case charInteger:
        value, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }
        return Integer(value), nil
    default:
        words := bytes.Split(header[:len(header)-2], []byte{' '})
        command := string(bytes.ToLower(words[0]))
        if bulkCommands[command] {
            blen, err := strconv.Atoi(string(words[len(words)-1]))
            if err != nil {
                return nil, err
            }

            data := make([]byte, blen)
            if err = readOrError(in, data); err != nil {
                return nil, err
            }

            if err = expectCrlf(in); err != nil {
                return nil, err
            }

            words[len(words)-1] = data
        }

        mbr := make(MultiBulkData, len(words))
        for i, word := range words {
            mbr[i] = BulkData(word)
        }
        return mbr, nil
    }
    return nil, newProtocolError(fmt.Sprintf("Unknown reply type: %#v (%#v)", header, string(header)))
}
