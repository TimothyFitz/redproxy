package main

import (
    "flag"
    "fmt"
    "net"
    "github.com/timothyfitz/redproxy"
)

type FrontendConn struct {
    *net.TCPConn
}

type BackendConn struct {
    *net.TCPConn
}

func copyRedis(from *net.TCPConn, to *net.TCPConn) error {
    for {
        v, err := redproxy.Read(from)
        if err != nil {
            return err
        }
        redproxy.Write(v, to)
    }
    return nil
}


func handleWrite(local FrontendConn, remote BackendConn) {
    // Handle Frontend to Backend communication
    //io.Copy(local, remote)
    copyRedis(local.TCPConn, remote.TCPConn)
    fmt.Println("io.Copy(local, remote) finished.")
    local.Close()
}

func handleRead(local FrontendConn, remote BackendConn) {
    // Handle Backend to Frontend communication
    //io.Copy(remote, local)
    copyRedis(remote.TCPConn, local.TCPConn)
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

    fe_conn := FrontendConn{local}
    be_conn := BackendConn{remote.(*net.TCPConn)}

    go handleWrite(fe_conn, be_conn)
    go handleRead(fe_conn, be_conn)
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

    for {
        conn, err := listener.AcceptTCP()
        if err != nil {
            panic(err)
        }
        go handleConn(conn)
    }
}
