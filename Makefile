.PHONY: build static

# Load the .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build:
	make prebuild
	go build -ldflags="-w -s" -o stack

build-amd:
	make prebuild
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o stack

prebuild:
	rm -rf static/*
	rm -rf dist/*
	yarn build
	cp -r assets/icons static/icons
	make build-py
	cp -r manage.py static .env.example dist

build-py:
	docker build -t $(PROJECT_NAME)-py-env -f build/dist.dockerfile .
	container_id=$$(docker create $(PROJECT_NAME)-py-env) && \
	docker cp $$container_id:/dist dist && \
	docker rm $$container_id
	mv dist/dist/* dist
	rm -r dist/dist

setup-project:
	go mod download
	python3 -m venv venv
	mkdir logs
	mkdir dist 
	python3 -m pip install --upgrade pip
	. venv/bin/activate; pip install -r requirements.txt
	cp .env.example .env
	. venv/bin/activate; python3 manage.py migrate
	yarn install --check-files
	yarn build
	. venv/bin/activate; python3 manage.py collectstatic --noinput
	make build
