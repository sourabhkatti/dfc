# from https://github.com/GoogleContainerTools/distroless/blob/main/examples/go/Dockerfile

FROM cgr.dev/ORG/go:1.22-dev AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM cgr.dev/ORG/static:latest

COPY --from=build /go/bin/app /
CMD ["/app"]
