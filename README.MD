## GoOut

这是一个可以将Tcp流量伪装成Http流量的程序,并可以提供HTTP代理端口和Sock5代理端口,支持全局流量代理转发和局部流量代理转发

GoOut客户端和GoOut服务端之间的传输的数据将会附加在HTTP POST里,作为一个正常的HTTP请求发送到GoOut服务端,由GoOut服务端转发到目的地。

同时GoOut客户端会提供SOCK5协议的代理端口、HTTP协议的代理端口。你可以使用系统自带的代理功能进行代理,无需下载任何工具。

## 特点

* 轻量小巧,极小的源代码
* 高性能,不对流量进行加密解密混淆,直接转发,极小的性能损耗

## 用法

### GoOut服务端

```
git clone https://github.com/blacknight2018/GoOut
cd GoOut/cmd/server
go build main.go
```

```
./main
```

### GoOut客户端

```
git clone https://github.com/blacknight2018/GoOut
cd GoOut/cmd/client
go build main.go
```

```
./main -h
```

```
GoOut version
Server 127.0.0.1:80
Usage of main.exe:
  -global
        开启全局模式
  -http
        使用Http代理协议,默认使用Sock5代理协议
  -server string
        GoOut服务端地址 (default "127.0.0.1:80")
```

> Http代理模式和Sock5代理模式默认端口都是7777

##### Http代理模式

```
main.exe -server=xxx.xxx.xxx.xxx:80  -http
```

##### Sock5代理模式

```
main.exe -server=xxx.xxx.xxx.xxx:80
```

### 在你的操作系统上使用GoOut客户端

> 图中打码地址即为你的GoOut客户端所在地址

###### 在Win10上使用GoOut客户端 (GoOut使用Http代理模式)

![avatar](images/Win10-http.png)

###### 在Android系统上使用GoOut客户端 (GoOut使用Http代理模式)

![avatar](images/Android-http.jpg)

###### 在Win10上使用Proxifier (GoOut使用Sock5代理模式)

![avatar](images/Win10-Proxifier.png)