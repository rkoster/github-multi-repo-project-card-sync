FROM golang:1.16
WORKDIR /go/src/github.com/rkoster/github-multi-repo-project-card-sync
RUN go get -d -v golang.org/x/net/html
ADD ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o sync .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/rkoster/github-multi-repo-project-card-sync/sync ./
CMD ["./sync"]
