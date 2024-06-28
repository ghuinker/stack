.PHONY: build static

# Load the .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build:
	docker build --platform linux/arm64 -t $(PROJECT_NAME) .

run:
	docker run -v ./db.sqlite3:/app/db.sqlite3 --env-file=.env -p 8000:80 $(PROJECT_NAME)

build-amd:
	docker build --platform linux/amd64 -t $(PROJECT_NAME)-amd .

run-prod:
	docker run -v -d ./db.sqlite3:/app/db.sqlite3 -v ./certs:/app/.local/share/certmagic --env-file=.env -p 80:80 -p 443:443 $(PROJECT_NAME)

save-image:
	docker save -o tmp/stack.tar $(PROJECT_NAME)-amd

load-image:
	docker load -i stack.tar

setup-project:
	go mod download
	python3 -m venv venv
	. venv/bin/activate; pip install -r requirements.txt
	cp .env.example .env
	. venv/bin/activate; python3 manage.py migrate
	yarn install --check-files
	yarn build
	. venv/bin/activate; python3 manage.py collectstatic --noinput
