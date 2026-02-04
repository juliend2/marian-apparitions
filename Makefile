.PHONY: run
run:
	go run .

.PHONY: clean
clean:
	rm -f data.sqlite3 marianapparitions

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o app query_helper.go init_db.go sorting.go main.go

.PHONY: deploy
deploy:
	cd ansible && ansible-playbook -i inventory.ini provision.yml

.PHONY: admin
admin:
	cd data_management/mariadmin/ && UV_ENV_FILE=mariadmin/.env uv run python manage.py runserver
