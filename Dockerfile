FROM golang:1.9 as builder

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

WORKDIR /go/src/github.com/dndungu/facade

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o facade .

FROM scratch

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

WORKDIR /bin

COPY --from=builder /go/src/github.com/dndungu/facade/facade .

ENV HTTP_PROXY_PORT 80

EXPOSE 80

ENV HTTPS_PROXY_PORT 443

EXPOSE 443

ENTRYPOINT ["/bin/facade"]
