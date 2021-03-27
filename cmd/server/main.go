package main

import (
	"bytes"
	"flag"
	"github.com/blacknight2018/GoOut/utils"
	"github.com/blacknight2018/GoProxys"
	"net"
	"time"
)

var limitTime *int64

func handleTCP(tcp *net.TCPConn) {
	var ioBuffer bytes.Buffer
	defer tcp.Close()
	var tcpWithTarget *net.TCPConn
	for {
		req, ok := utils.ParseHttpRequest(tcp, &ioBuffer)
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
			_, err = utils.WriteHttpResponse(tcp, []byte("Done"))
			if err != nil {
				return
			}

			//Recv from remote
			go func(target *net.TCPConn, proxyClient *net.TCPConn) {
				for {
					var buff [10048576]byte
					target.SetReadDeadline(time.Now().Add(time.Second * 300))
					n, err := target.Read(buff[:])
					if err != nil {
						target.Close()
						return
					}
					n, err = utils.WriteHttpResponse(proxyClient, buff[:n])
					if err != nil {
						target.Close()
						return
					}
				}
			}(tcpWithTarget, tcp)
		} else if path == "/send" {
			_, err := tcpWithTarget.Write(req.Body)
			if err != nil {
				tcpWithTarget.Close()
				return
			}
		} else if path == "/" {
			_, err := utils.WriteHttpResponseWithCt(tcp, []byte("Hello,GFW"), "text/plain; charset=utf-8")
			if err != nil {
				tcpWithTarget.Close()
				return
			}
			return
		}
	}
}
func main() {
	GoProxys.StartWatch()
	//limitTime = flag.Int64("time", 20, "Tcp read limit time (second)")
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
