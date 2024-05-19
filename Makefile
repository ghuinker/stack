.PHONY: build static

prebuild:
	rm -rf static/*
	python3 manage.py collectstatic --no-input
	yarn build
	cp -r assets/icons static/icons
	find app -type d -name '__pycache__' -exec rm -r {} +

build:
	make prebuild
	go build -ldflags="-w -s" -o stack

build-amd:
	make prebuild
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o stack

setup-project:
	cp .env.example .env
	python3 -m venv venv
	python3 -m pip install --upgrade pip
	. venv/bin/activate; pip install -r requirements.txt
	./scripts/update_secret_key.sh
	. venv/bin/activate; python3 manage.py migrate
	yarn install --check-files
	yarn build
	make static
