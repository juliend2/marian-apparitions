.PHONY: run
run:
	OTEL_EXPORTER_OTLP_INSECURE=true go run .

.PHONY: clean
clean:
	rm -f data.sqlite3 marianapparitions

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o app query_helper.go init_db.go sorting.go main.go

.PHONY: deploy
deploy: build
	cd ansible && ansible-playbook -i inventory.ini provision.yml

.PHONY: admin
admin:
	cd data_management/mariadmin/ && UV_ENV_FILE=mariadmin/.env uv run python manage.py runserver

.PHONY: shell
shell:
	cd data_management/mariadmin/ && UV_ENV_FILE=mariadmin/.env uv run python manage.py shell

.PHONY: journal
journal:
	journalctl -f -u marianapparitions-*.service

.PHONY: docker
docker:
	docker compose up
