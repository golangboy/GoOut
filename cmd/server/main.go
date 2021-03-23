package main

import (
	"flag"
	"github.com/blacknight2018/GoOut/utils"
	"net"
	"time"
)

var limitTime *int64

func handleTCP(tcp *net.TCPConn) {
	defer tcp.Close()
	var tcpWithTarget *net.TCPConn
	for {
		req, ok := utils.ParseHttpRequest(tcp, time.Minute*(time.Duration(*limitTime)))
		if !ok {
			return
		}
		path := req.Url
		if path == "/conn" {
			targetHost := string(req.Body)
			tcpAddr, err := net.ResolveTCPAddr("tcp4", targetHost)
			if err != nil {
				return
			}

			//repeat connect
			if tcpWithTarget != nil {
				tcpWithTarget.Close()
				return
			}
			tcpWithTarget, err = net.DialTCP("tcp4", nil, tcpAddr)
			if err != nil {
				return
			}
			utils.WriteHttpResponse(tcp, []byte("Done"))

			//Recv from remote
			go func(target *net.TCPConn) {
				for {
					var buff [1024]byte
					n, err := target.Read(buff[:])
					if err != nil {
						target.Close()
						return
					}
					utils.WriteHttpResponse(tcp, buff[:n])
				}
			}(tcpWithTarget)
		} else if path == "/send" {
			tcpWithTarget.Write(req.Body)
		} else if path == "/" {
			utils.WriteHttpResponseWithCt(tcp, []byte("Hello,GFW"), "text/plain; charset=utf-8")
			return
		}
	}
}
func main() {

	limitTime = flag.Int64("time", 1, "Tcp read limit time (minute)")
	flag.Parse()

	ta, err := net.ResolveTCPAddr("tcp4", ":80")
	tc, err := net.ListenTCP("tcp4", ta)
	if err != nil {
		panic(err)
	}

	for {
		client, err := tc.AcceptTCP()
		if err == nil && client != nil {
			go handleTCP(client)
		} else if client != nil {
			client.Close()
		}
	}
}
