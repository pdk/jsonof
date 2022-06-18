all: jsonf jsonf.lnx

clean:
	rm jsonf jsonf.lnx

jsonf: jsonf.go
	go build -o jsonf jsonf.go

jsonf.lnx: jsonf.go
	env GOOS=linux GOARCH=amd64 go build -o jsonf.lnx jsonf.go
