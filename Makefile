all: jsonobj jsonobj.lnx

clean:
	rm jsonobj jsonobj.lnx

jsonobj: jsonobj.go
	go build -o jsonobj jsonobj.go

jsonobj.lnx: jsonobj.go
	env GOOS=linux GOARCH=amd64 go build -o jsonobj.lnx jsonobj.go
