FROM golang:1.9 as builder

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

WORKDIR /go/src/github.com/dndungu/director

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o director .

FROM alpine:latest

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

WORKDIR /bin

COPY --from=builder /go/src/github.com/dndungu/director/director .

RUN apk --update add ca-certificates

ENV HTTP_PROXY_PORT 80

EXPOSE 80

ENV HTTPS_PROXY_PORT 443

EXPOSE 443

ENTRYPOINT ["/bin/director"]
