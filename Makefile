all: jsonof jsonof.lnx

clean:
	rm jsonof jsonof.lnx

jsonof: jsonof.go
	go build -o jsonof jsonof.go

jsonof.lnx: jsonof.go
	env GOOS=linux GOARCH=amd64 go build -o jsonof.lnx jsonof.go
