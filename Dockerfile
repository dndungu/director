FROM golang:1.11 as builder

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

WORKDIR /go/src/github.com/dndungu/director

COPY . .

ARG COMMIT_SHA

RUN go get -u github.com/golang/dep/cmd/dep

RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o director -ldflags "-X main.CommitSha=${COMMIT_SHA}" .

FROM scratch

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

COPY --from=builder /etc/ssl /etc/ssl

WORKDIR /bin

COPY --from=builder /go/src/github.com/dndungu/director/director .

ENV HTTP_PROXY_PORT 80

EXPOSE 80

ENV HTTPS_PROXY_PORT 443

EXPOSE 443

ENTRYPOINT ["/bin/director"]
