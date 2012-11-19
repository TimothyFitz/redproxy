package redproxy

import (
    "flag"
    "fmt"
    "net"
    //"io"
    "github.com/garyburd/redigo/redis"
)

type FrontendConn struct {
    *net.TCPConn
}

type BackendConn struct {
    redis.Conn
}

func handleWrite(local FrontendConn, remote BackendConn) {
    // Handle Frontend to Backend communication
    //io.Copy(local, remote)
    fmt.Println("io.Copy(local, remote) finished.")
    local.Close()
}

func handleRead(local FrontendConn, remote BackendConn) {
    // Handle Backend to Frontend communication
    //io.Copy(remote, local)
    fmt.Println("io.Copy(remote, local) finished.")
    remote.Close()
}

func handleConn(local *net.TCPConn) {
    remote, err := redis.Dial("tcp", *remote_addr)

    fmt.Println("New connection")

    if remote == nil {
        fmt.Printf("remote dial failed: %v\n", err)

        return
    }

    fe_conn := FrontendConn{local}
    be_conn := BackendConn{remote}

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
