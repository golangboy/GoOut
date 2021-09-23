package api

import (
	"bytes"
	"github.com/blacknight2018/GoOut/utils"
	"net"
)

// The current connection(conn) will translate data stream through the GoOut Server
func TcpOnProxy(conn net.Conn, laddr *net.TCPAddr, host string, port string, GoOutServer *string) {
	var ioBuffer bytes.Buffer
	//Connect to web proxy server
	tcpAddr, err := net.ResolveTCPAddr("tcp4", *GoOutServer)
	if err != nil {
		conn.Close()
		return
	}
	webTcp, err := net.DialTCP("tcp4", laddr, tcpAddr)
	if err != nil {
		conn.Close()
		return
	}
	//Request connect target
	utils.WriteHttpRequest(webTcp, "/conn", []byte(host+":"+port))
	req, ok := utils.ParseHttpResponse(webTcp, &ioBuffer)
	if !ok {
		conn.Close()
		webTcp.Close()
		return
	}
	//Connect finish
	if string(req.Body) != "Done" {
		conn.Close()
		webTcp.Close()
		return
	}

	//Recv from client,send to web server
	go func(client net.Conn, webSer *net.TCPConn) {
		for {
			var buff [10485]byte
			//client.SetReadDeadline(time.Now().Add(time.Second * 300))
			n, err := client.Read(buff[:])
			if err != nil {
				client.Close()
				webTcp.Close()
				return
			}
			utils.WriteHttpRequest(webSer, "/send", buff[:n])
		}
	}(conn, webTcp)

	//Recv from web server, send to local proxy client
	for {
		req, ok = utils.ParseHttpResponse(webTcp, &ioBuffer)
		if !ok {
			webTcp.Close()
			conn.Close()
			return
		}
		conn.Write(req.Body)
	}
}
