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

save-image:
	docker save -o tmp/stack.tar $(PROJECT_NAME)-amd


setup-project:
	go mod download
	python3 -m venv venv
	. venv/bin/activate; pip install -r requirements.txt
	cp .env.example .env
	sed -i '' '1s/.*/DEBUG=True/' .env
	. venv/bin/activate; python3 manage.py migrate
	yarn install --check-files
	yarn build
	. venv/bin/activate; python3 manage.py collectstatic --noinput

