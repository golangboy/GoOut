package main

import (
	"flag"
	"fmt"
	"github.com/blacknight2018/GoOut/utils"
	"github.com/blacknight2018/GoProxys"
	"github.com/oschwald/geoip2-golang"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

var server *string
var limitTime *int64
var httpMode *bool
var global *bool
var geoDb *geoip2.Reader

func ioCopyWithTimeOut(dst net.Conn, src net.Conn, timeOut time.Duration) {
	var buff [10240]byte
	for {
		src.SetReadDeadline(time.Now().Add(timeOut))
		n, err := src.Read(buff[:])
		if err != nil {
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
	go ioCopyWithTimeOut(remote, conn, time.Minute*(time.Duration(*limitTime)))
	ioCopyWithTimeOut(conn, remote, time.Minute*(time.Duration(*limitTime)))

}
func OnProxy(conn net.Conn, host string, port string) {
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

func StartHttpProxyServer() {
	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", ":7777")
	s := GoProxys.DefaultHttp()
	s.HttpConnect = func(conn net.Conn, host string, port string) {
		ip := utils.GetFirstIpByHost(host)
		if *global {
			OnProxy(conn, host, port)
			return
		}
		if geoDb != nil {
			record, err := geoDb.City(net.ParseIP(ip))
			if err == nil && record.Country.Names["en"] != "China" {
				fmt.Println(host, record.Country.Names["en"])
				OnProxy(conn, host, port)
				return
			}
		}
		OnDirect(conn, host, port)
	}
	s.RunHttpProxy(b)
}

func StartSock5ProxyServer() {
	go GoProxys.StartWatch()
	b, _ := net.ResolveTCPAddr("tcp4", ":7777")
	s := GoProxys.DefaultSocket5()
	s.TcpConnect = func(conn net.Conn, host string, port string) {
		ip := utils.GetFirstIpByHost(host)
		if *global {
			OnProxy(conn, host, port)
			return
		}
		if geoDb != nil {
			record, err := geoDb.City(net.ParseIP(ip))
			if err == nil && record.Country.Names["en"] != "China" {
				fmt.Println(host, record.Country.Names["en"])
				OnProxy(conn, host, port)
				return
			}
		}
		OnDirect(conn, host, port)
	}
	s.RunSocket5Proxy(b)
}

func downLoadGeoLite2() {
	fmt.Println("downloading GeoIp2 db")
	fileName := "GeoLite2-City.mmdb"
	_, err := os.Stat(fileName)
	if err == nil {
		return
	}
	url := `https://raw.githubusercontent.com/blacknight2018/GeoLite2/master/GeoLite2-City.mmdb`
	resp, err := http.Get(url)
	if err == nil {
		fs, _ := os.Create(fileName)
		io.Copy(fs, resp.Body)
	}
}
func main() {
	server = flag.String("server", "127.0.0.1:80", "GoOut服务端地址")
	limitTime = flag.Int64("time", 1, "TCP连接超时时间(分钟)")
	httpMode = flag.Bool("http", false, "使用Http代理协议,默认false,即默认使用Sock5代理协议")
	global = flag.Bool("global", false, "是否开启全局模式,默认false,即默认只有国外流量走代理")
	flag.Parse()
	if server == nil || len(*server) == 0 {
		return
	}
	if false == *global {
		downLoadGeoLite2()
		geoDb, _ = geoip2.Open("GeoLite2-City.mmdb")
	}
	if *httpMode {
		StartHttpProxyServer()
	} else {
		StartSock5ProxyServer()
	}
}
