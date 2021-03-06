package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/blacknight2018/GoOut/api"
	"github.com/blacknight2018/GoOut/utils"
	"github.com/blacknight2018/GoProxys"
	"github.com/oschwald/geoip2-golang"
	"net"
	"time"
)

var server *string
var limitTime *int64
var httpMode *bool
var global *bool
var geoDb *geoip2.Reader

func ioCopyWithTimeOut(dst net.Conn, src net.Conn, timeOut time.Duration) {
	var buff [10485]byte
	for {
		//src.SetReadDeadline(time.Now().Add(timeOut))
		n, err := src.Read(buff[:])
		if err != nil {
			dst.Close()
			src.Close()
			return
		}
		dst.Write(buff[:n])
	}
}
func OnDirect(conn net.Conn, host string, port string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host+":"+port)
	if err != nil {
		conn.Close()
		return
	}
	remote, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		conn.Close()
		return
	}
	go ioCopyWithTimeOut(remote, conn, time.Second*(10))
	ioCopyWithTimeOut(conn, remote, time.Second*(10))

}
func OnProxy(conn net.Conn, host string, port string) {
	var ioBuffer bytes.Buffer
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

var listenPort *string

func StartHttpProxyServer() {
	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", *listenPort)
	s := GoProxys.DefaultHttp()
	s.HttpConnect = func(conn net.Conn, host string, port string) {
		ip := utils.GetFirstIpByHost(host)
		if *global {
			//OnProxy(conn, host, port)
			api.TcpOnProxy(conn, nil, host, port, server)
			return
		}

		if false == utils.IsChinaIP(ip) {
			fmt.Println(host + " are not in china,through proxy")
			api.TcpOnProxy(conn, nil, host, port, server)
			return
		}

		OnDirect(conn, host, port)
	}
	s.RunHttpProxy(b)
}

func StartSock5ProxyServer() {
	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", *listenPort)
	s := GoProxys.DefaultSocket5()
	s.TcpConnect = func(conn net.Conn, host string, port string) {
		ip := utils.GetFirstIpByHost(host)
		if *global {
			fmt.Println(host)
			api.TcpOnProxy(conn, nil, host, port, server)
			return
		}
		if false == utils.IsChinaIP(ip) {
			fmt.Println(host + " are not in china,through proxy")
			api.TcpOnProxy(conn, nil, host, port, server)
			return
		}
		OnDirect(conn, host, port)
	}
	s.RunSocket5Proxy(b)
}

//func downLoadGeoLite2() {
//	fmt.Println("downloading GeoIp2 db")
//	fileName := "GeoLite2-City.mmdb"
//	_, err := os.Stat(fileName)
//	if err == nil {
//		return
//	}
//	url := `https://raw.githubusercontent.com/blacknight2018/GeoLite2/master/GeoLite2-City.mmdb`
//	resp, err := http.Get(url)
//	if err == nil {
//		fs, _ := os.Create(fileName)
//		io.Copy(fs, resp.Body)
//	}
//}

var bdServer, bdVersion string

func main() {
	if len(bdServer) == 0 {
		bdServer = "127.0.0.1:80"
	}
	fmt.Println("GoOut version", bdVersion)
	fmt.Println("Server", bdServer)
	server = flag.String("server", bdServer, "GoOut???????????????")
	listenPort = flag.String("port", ":7777", "?????????????????????,?????? :7777  ??????7777??????")
	//limitTime = flag.Int64("time", 20, "TCP??????????????????(???)")
	httpMode = flag.Bool("http", false, "??????Http????????????,????????????Sock5????????????")
	global = flag.Bool("global", false, "????????????????????????")
	flag.Parse()
	if server == nil || len(*server) == 0 {
		return
	}

	fmt.Println("GoOut?????????:" + "[" + *server + "]" + " " + "?????????????????????[" + *listenPort + "]")
	if *httpMode {
		StartHttpProxyServer()
	} else {
		StartSock5ProxyServer()
	}
}
