package main

import (
    "fmt"
    "flag"
    "net"
    "io"
)

func handleConn(local *net.TCPConn) {
    remote, err := net.Dial("tcp", *remote_addr)
    if remote == nil {
        fmt.Printf("remote dial failed: %v\n", err)
        return
    }
    go io.Copy(local, remote)
    go io.Copy(remote, local)
}

var port_str *string = flag.String("p", "9999", "local port")
var remote_addr *string = flag.String("r", "localhost:6379", "remote address")


func main() {
    flag.Parse()

    fmt.Printf("Listening on port %v\nProxying: %v\n\n", *port_str, *remote_addr)

    addr, err := net.ResolveTCPAddr("tcp", "localhost:" + *port_str)
    if err != nil { panic(err) }

    listener, err := net.ListenTCP("tcp", addr)
    if err != nil { panic(err) }

    for {
        conn, err := listener.AcceptTCP()
        if err != nil {
            panic(err)
        }
        go handleConn(conn)
    }
}