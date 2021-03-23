package utils

import (
	"net"
	"strconv"
	"strings"
	"time"
)

type httpReqHeader struct {
	Method  string
	Url     string
	Version string
	Headers []string
	Raw     string
	Body    []byte
}

func ReadStringUntil(reader net.Conn, token string, timeOut time.Duration) (string, bool) {
	var t [1]byte
	var pos int
	var tmp string
	for {
		reader.SetReadDeadline(time.Now().Add(timeOut))
		_, err := reader.Read(t[:])
		tmp = tmp + string(t[:])
		if err != nil {
			return "", false
		}
		if t[0] == token[pos] {
			pos = pos + 1
		} else {
			pos = 0
		}
		if pos == len(token) {
			break
		}

	}
	return tmp[:len(tmp)-len(token)], true
}
func GetHttpKey(s string) string {
	r := strings.Split(s, ": ")
	if len(r) != 2 {
		return ""
	}
	return r[0]
}
func GetHttpValue(s string) string {
	r := strings.Split(s, ": ")
	if len(r) != 2 {
		return ""
	}
	return r[1]
}
func ParseHttpRequest(reader net.Conn, timeOut time.Duration) (httpReqHeader, bool) {
	var ret httpReqHeader
	var buff = make([]byte, 1024)
	method, ok := ReadStringUntil(reader, " ", timeOut)
	if !ok {
		reader.SetReadDeadline(time.Now().Add(timeOut))

		reader.Read(buff[:])
		ret.Raw = ret.Raw + string(buff[:])
		return ret, false
	}
	ret.Method = method
	ret.Raw = ret.Raw + ret.Method + " "

	u, ok := ReadStringUntil(reader, " ", timeOut)

	//http://xxx.xxx/abc -> /abc
	if len(u) >= 7 && u[:4] == "http" {
		u = u[7:]
		pos := strings.Index(u, "/")
		if pos != -1 {
			u = u[pos:]
		}
	}
	if !ok {
		reader.SetReadDeadline(time.Now().Add(time.Second * 3 * 60))
		reader.Read(buff[:])
		ret.Raw = ret.Raw + string(buff[:])
		return ret, false
	}
	ret.Raw = ret.Raw + u + " "
	ret.Url = u
	v, ok := ReadStringUntil(reader, "\r\n", timeOut)
	if !ok {
		reader.SetReadDeadline(time.Now().Add(time.Second * 3 * 60))
		reader.Read(buff[:])
		ret.Raw = ret.Raw + string(buff[:])
		return ret, false
	}
	ret.Raw = ret.Raw + v + "\r\n"
	ret.Version = v
	var needRead int
	for {
		head, ok := ReadStringUntil(reader, "\r\n", timeOut)
		ret.Raw = ret.Raw + head + "\r\n"
		if !ok {
			break
		}
		if len(head) == 0 {
			break
		}
		if strings.ToLower(GetHttpKey(head)) == "content-length" {
			tmp := GetHttpValue(head)
			needRead, _ = strconv.Atoi(tmp)
		}
		ret.Headers = append(ret.Headers, head)
	}
	for {
		var buff [1024]byte
		var n int
		var err error
		if needRead > 1024 {
			reader.SetReadDeadline(time.Now().Add(time.Second * 3 * 60))
			n, err = reader.Read(buff[:])
			needRead = needRead - 1024
		} else {
			reader.SetReadDeadline(time.Now().Add(time.Second * 3 * 60))
			n, err = reader.Read(buff[:needRead])
			needRead = needRead - n
		}
		if err != nil {
			break
		}
		ret.Body = append(ret.Body, buff[:n]...)
		if needRead == 0 {
			break
		}
	}
	return ret, true
}
func WriteHttpRequest(tcp *net.TCPConn, path string, data []byte) {
	payload := "POST XXX HTTP/1.1\r\nConnection: keep-alive\r\nContent-Length: YYY\r\nContent-Type: application/octet-stream\r\n\r\n"
	payload = strings.ReplaceAll(payload, "XXX", path)
	payload = strings.ReplaceAll(payload, "YYY", strconv.Itoa(len(data)))
	tcp.Write([]byte(payload))
	tcp.Write(data)
}
func WriteHttpResponse(tcp *net.TCPConn, data []byte) {
	payload := "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: xxx\r\n\r\n"
	payload = strings.ReplaceAll(payload, "xxx", strconv.Itoa(len(data)))
	tcp.Write([]byte(payload))
	tcp.Write(data)
}
func WriteHttpResponseWithCt(tcp *net.TCPConn, data []byte, contentType string) {
	payload := "HTTP/1.1 200 OK\r\nContent-Type: yyy\r\nContent-Length: xxx\r\n\r\n"
	payload = strings.ReplaceAll(payload, "xxx", strconv.Itoa(len(data)))
	payload = strings.ReplaceAll(payload, "yyy", contentType)
	tcp.Write([]byte(payload))
	tcp.Write(data)
}

func GetFirstIpByHost(host string) string {
	ip, err := net.LookupIP(host)
	if err == nil && len(ip) > 0 {
		return ip[0].String()
	}
	return ""
}
