package main

import (
    "fmt"
    "flag"
    "net"
    "io"
)

func handleWrite(local *net.TCPConn, remote *net.TCPConn) {
    io.Copy(local, remote)
    fmt.Println("io.Copy(local, remote) finished.")
    local.Close()
}

func handleRead(local *net.TCPConn, remote *net.TCPConn) {
    io.Copy(remote, local)
    fmt.Println("io.Copy(remote, local) finished.")
    remote.Close()
}

func handleConn(local *net.TCPConn) {
    remote, err := net.Dial("tcp", *remote_addr)

    fmt.Println("New connection")

    if remote == nil {
        fmt.Printf("remote dial failed: %v\n", err)

        return
    }
    go handleWrite(local, remote.(*net.TCPConn))
    go handleRead(local, remote.(*net.TCPConn))
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