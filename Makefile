run:
	go run .

clean:
	rm -f data.sqlite3 marianapparitions

build:
	GOOS=linux GOARCH=amd64 go build -o app query_helper.go main.go

deploy:
	cd ansible && ansible-playbook -i inventory.ini provision.yml