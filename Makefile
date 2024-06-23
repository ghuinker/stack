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
	rm -rf dist
	docker build -t $(PROJECT_NAME)-build-env -f build/build.dockerfile .
	container_id=$$(docker create $(PROJECT_NAME)-build-env) && \
	docker cp $$container_id:/dist dist && \
	docker rm $$container_id

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
