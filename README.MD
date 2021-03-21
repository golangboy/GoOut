### GoOut

A proxy program,base on Sock5,use http body as payload translate data.

### Usage

#### In server

```shell
git clone https://github.com/blacknight2018/GoOut
cd GoOut/cmd/server
go build main.go
./main
```

#### In client

```shell
git clone https://github.com/blacknight2018/GoOut
cd GoOut/cmd/client
go build main.go
```

```shell
./main -h
```

```shell
Usage of main.exe:
  -server string
        This GoOut server address
  -time int
        Tcp read limit time (minute) (default 60)

```

```shell
main.exe -server=xxx.xxx.xxx.xxx:80
```

