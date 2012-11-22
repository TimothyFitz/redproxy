package main

import (
    "flag"
    "fmt"
    "github.com/timothyfitz/redproxy"
    "net"
)

type FrontendConn struct {
    *net.TCPConn
}

type BackendConn struct {
    *net.TCPConn
}

type Response struct {
    value *interface{}
}

type Request struct {
    promise chan<- Response
    value   *interface{}
}

func handleBackend(requests chan Request, remote BackendConn) {
    promises := make(chan chan<- Response)
    go handleBackendResponses(promises, remote)
    for request := range requests {
        promises <- request.promise
        redproxy.Write(*request.value, remote.TCPConn)
    }
}

func handleBackendResponses(promises chan chan<- Response, remote BackendConn) {
    for promise := range promises {
        v, err := redproxy.Read(remote)
        if err != nil {
            // TODO: Do the right thing here (EOF vs Other)
        }
        promise <- Response{&v}
    }
}

func handleFrontend(requests chan Request, remote FrontendConn) {
    responses := make(chan Response)
    defer close(responses)
    defer remote.Close()

    go handleFrontendResponses(responses, remote)

    for {
        v, err := redproxy.Read(remote.TCPConn)
        if err != nil {
            // TODO: Do the right thing here (EOF vs Other)
            return
        }
        requests <- Request{responses, &v}
    }
}

func handleFrontendResponses(responses chan Response, remote FrontendConn) {
    for response := range responses {
        redproxy.Write(*response.value, remote.TCPConn)
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
