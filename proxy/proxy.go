package main

import (
    "flag"
    "fmt"
    "github.com/timothyfitz/redproxy"
    "net"
    "bufio"
    "log"
    "io"
    //"runtime"
)

const BUFFER_SIZE = 100

type FrontendConn struct {
    *net.TCPConn
}

type BackendConn struct {
    *net.TCPConn
}

type Tag uint64

type Response struct {
    value *interface{}
    id Tag
}

type Promise struct {
    promise chan<- Response
    id Tag
}

type Request struct {
    Promise
    value *interface{}
}

func handleBackend(requests chan Request, remote BackendConn) {
    promises := make(chan Promise, BUFFER_SIZE)
    output := bufio.NewWriter(remote.TCPConn)
    go handleBackendResponses(promises, remote)
    for request := range requests {
        promises <- request.Promise
        redproxy.Write(*request.value, output)
        output.Flush()
    }
}

func handleBackendResponses(promises <-chan Promise, remote BackendConn) {
    in := bufio.NewReader(remote)
    for promise := range promises {
        v, err := redproxy.Read(in)
        if err != nil {
            // TODO: Do the right thing here (EOF vs Other)
            log.Panicln("BE_ERR:", err)
        }
        promise.promise <- Response{&v, promise.id}
    }
}

func handleFrontend(requests chan Request, remote FrontendConn) {
    request_id := Tag(1)
    responses := make(chan Response, BUFFER_SIZE)
    defer close(responses)
    defer remote.Close()

    go handleFrontendResponses(responses, remote)

    in := bufio.NewReader(remote)

    for {
        v, err := redproxy.Read(in)
        if err != nil {
            // TODO: Do the right thing here (EOF vs Other)
            if err == io.EOF {
                fmt.Println("Connection closed.")
            } else {
                fmt.Printf("Error while reading remote conn: %v\n", err)
            }
            return
        }

        requests <- Request{Promise{responses, request_id}, &v}
        request_id++
    }
}

func handleFrontendResponses(responses chan Response, remote FrontendConn) {
    request_id := Tag(1)
    queued_responses := make(map[Tag] Response)
    output := bufio.NewWriter(remote)
    for response := range responses {
        queued_responses[response.id] = response

        for {
            response, ok := queued_responses[request_id]
            if !ok {
                break
            }
            delete(queued_responses, request_id)
            redproxy.Write(*response.value, output)
            output.Flush()
            request_id++
        }
    }
}

func handleConn(requests chan Request, local *net.TCPConn) {
    fmt.Println("New connection")

    fe_conn := FrontendConn{local}

    go handleFrontend(requests, fe_conn)
}

var port_str *string = flag.String("p", "9999", "local port")
var remote_addr *string = flag.String("r", "localhost:6379", "remote address")

func main() {
    //runtime.GOMAXPROCS(runtime.NumCPU())

    flag.Parse()

    fmt.Printf("Listening on port %v\nProxying: %v\n\n", *port_str, *remote_addr)

    addr, err := net.ResolveTCPAddr("tcp", "localhost:"+*port_str)
    if err != nil {
        panic(err)
    }

    listener, err := net.ListenTCP("tcp", addr)
    if err != nil {
        panic(err)
    }

    remote, err := net.Dial("tcp", *remote_addr)

    if remote == nil {
        fmt.Printf("remote dial failed: %v\n", err)

        return
    }

    be_conn := BackendConn{remote.(*net.TCPConn)}

    requests := make(chan Request)

    go handleBackend(requests, be_conn)

    for {
        conn, err := listener.AcceptTCP()
        if err != nil {
            panic(err)
        }
        go handleConn(requests, conn)
    }
}
