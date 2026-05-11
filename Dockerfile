ARG NODE_VERSION=22
ARG GO_VERSION=1.23

FROM node:${NODE_VERSION}-alpine AS css
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY view/ ./view/
RUN npx @tailwindcss/cli -i view/css/app.css -o /styles.css

FROM golang:${GO_VERSION}-alpine AS build

ENV CGO_ENABLED=0

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=css /styles.css ./public/styles.css

RUN go install github.com/a-h/templ/cmd/templ@v0.3.906
RUN templ generate

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w" \
    -o /app ./main.go

FROM gcr.io/distroless/static@sha256:d6fa9db9548b5772860fecddb11d84f9ebd7e0321c0cb3c02870402680cc315f AS final

USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /app /app
COPY --chown=nonroot:nonroot ./content /content
COPY --from=build --chown=nonroot:nonroot /src/public /public

ENTRYPOINT ["./app"]
