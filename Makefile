run:
	go run .

clean:
	rm -f data.sqlite3 marianapparitions

build:
	GOOS=linux GOARCH=amd64 go build -o app main.go
