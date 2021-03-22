package main

import (
	"flag"
	"github.com/blacknight2018/GoOut/utils"
	"github.com/blacknight2018/GoProxys"
	"net"
	"time"
)

var server *string
var limitTime *int64
var httpMode *bool

func StartHttpProxyServer() {

	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", ":7777")
	s := GoProxys.DefaultHttp()
	s.HttpConnect = func(conn net.Conn, host string, port string) {
		//Connect to web proxy server
		tcpAddr, err := net.ResolveTCPAddr("tcp4", *server)
		if err != nil {
			conn.Close()
			return
		}
		webTcp, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			conn.Close()
			return
		}
		//Request connect target
		utils.WriteHttpRequest(webTcp, "/conn", []byte(host+":"+port))
		req, ok := utils.ParseHttpRequest(webTcp, time.Minute*(time.Duration(*limitTime)))
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
				var buff [1024]byte
				client.SetReadDeadline(time.Now().Add(time.Minute * (time.Duration(*limitTime))))
				n, err := client.Read(buff[:])
				if err != nil {
					client.Close()
					return
				}
				utils.WriteHttpRequest(webSer, "/send", buff[:n])
			}
		}(conn, webTcp)

		//Recv from web server, send to local proxy client
		for {
			req, ok = utils.ParseHttpRequest(webTcp, time.Minute*(time.Duration(*limitTime)))
			if !ok {
				webTcp.Close()
				conn.Close()
				return
			}
			conn.Write(req.Body)
		}
	}
	s.RunHttpProxy(b)
}

func StartSock5ProxyServer() {
	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", ":7777")
	s := GoProxys.DefaultSocket5()
	s.TcpConnect = func(conn net.Conn, host string, port string) {
		//Connect to web proxy server
		tcpAddr, err := net.ResolveTCPAddr("tcp4", *server)
		if err != nil {
			conn.Close()
			return
		}
		webTcp, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			conn.Close()
			return
		}
		//Request connect target
		utils.WriteHttpRequest(webTcp, "/conn", []byte(host+":"+port))
		req, ok := utils.ParseHttpRequest(webTcp, time.Minute*(time.Duration(*limitTime)))
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
				var buff [1024]byte
				client.SetReadDeadline(time.Now().Add(time.Minute * (time.Duration(*limitTime))))
				n, err := client.Read(buff[:])
				if err != nil {
					client.Close()
					return
				}
				utils.WriteHttpRequest(webSer, "/send", buff[:n])
			}
		}(conn, webTcp)

		//Recv from web server, send to local proxy client
		for {
			req, ok = utils.ParseHttpRequest(webTcp, time.Minute*(time.Duration(*limitTime)))
			if !ok {
				webTcp.Close()
				conn.Close()
				return
			}
			conn.Write(req.Body)
		}
	}
	s.RunSocket5Proxy(b)
}
func main() {
	server = flag.String("server", "127.0.0.1:80", "GoOut服务端地址")
	limitTime = flag.Int64("time", 1, "TCP连接超时时间(分钟)")
	httpMode = flag.Bool("http", false, "使用Http代理协议,默认false，即默认使用Sock5代理协议")
	flag.Parse()
	if server == nil || len(*server) == 0 {
		return
	}
	if *httpMode {
		StartHttpProxyServer()
	} else {
		StartSock5ProxyServer()
	}

}
