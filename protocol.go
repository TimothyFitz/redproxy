package redproxy

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strconv"
)

type MultiBulkData []BulkData
type BulkData []byte
type SingleLine []byte
type ErrorMessage []byte
type Integer int64

const (
    charMultyBulkData = '*'
    charBulkData      = '$'
    charSingleLine    = '+'
    charErrorMessage  = '-'
    charInteger       = ':'
)

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
    case charBulkData:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }

        if length < 0 {
            return nil, nil
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

        return BulkData(data), nil

    case charMultyBulkData:
        length, err := strconv.Atoi(string(msg_body))
        if err != nil {
            return nil, err
        }
        mbd := make(MultiBulkData, length)
        for i := 0; i < length; i++ {
            v, err := read(in)
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
    }
    return nil, newProtocolError(fmt.Sprintf("Unknown reply type: %#v (%#v)", header, string(header)))
}

func Read(in io.Reader) (v interface{}, err error) {
    v, err = read(bufio.NewReader(in))
    return
}
