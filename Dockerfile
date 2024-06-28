FROM --platform=$BUILDPLATFORM node:20.14 as assetbuild
WORKDIR /build
COPY package.json yarn.lock /build
RUN yarn install
COPY postcss.config.js tailwind.config.js vite.config.js /build
COPY assets /build/assets
# Required for tailwindcss to properly purge
COPY app/templates /build/app/templates
RUN yarn build
RUN cp -r /build/assets/icons static/icons

FROM --platform=$TARGETPLATFORM python:3.12-alpine as pybase
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY .env.example .env
COPY manage.py .
COPY app /app/app

FROM pybase as pystatic
RUN DEBUG=True PYTHONPATH=/build python3 manage.py collectstatic

from --platform=$BUILDPLATFORM golang:1.21 as gobuild
ARG TARGETARCH
ARG TARGETOS
WORKDIR /build
COPY go.mod go.sum .
RUN go mod download
COPY main.go .
COPY --from=assetbuild /build/static /build/static
COPY --from=pystatic /app/static /build/static
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o stack

FROM pybase as final
ARG TARGETPLATFORM
WORKDIR /app
RUN apk add libc6-compat

COPY --from=assetbuild /build/static/assets/manifest.json static/assets/manifest.json
COPY --from=gobuild /build/stack /app/stack
CMD ["./stack"]
