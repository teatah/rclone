FROM golang:1.24-alpine AS build_stage

WORKDIR /go/src/redditclone_app
COPY . .
RUN go mod download
RUN go build -o redditclone ./cmd/redditclone

FROM alpine AS run_stage

COPY --from=build_stage /go/src/redditclone_app /redditclone_app_binary
WORKDIR /redditclone_app_binary
RUN chmod +x .
EXPOSE 8080/tcp
ENTRYPOINT [ "./redditclone" ]