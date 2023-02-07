FROM golang:1.20-rc
WORKDIR /go/src/github.com/rkoster/github-multi-repo-project-card-sync
ADD ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o sync .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/rkoster/github-multi-repo-project-card-sync/sync ./
CMD ["./sync"]
