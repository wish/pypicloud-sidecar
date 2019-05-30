FROM golang:1.12
# RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/wish/pypicloud-sidecar/
COPY . /go/src/github.com/wish/pypicloud-sidecar/
# RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .




FROM alpine:3.7
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/wish/pypicloud-sidecar/pypicloud-sidecar /root/sidecar
CMD /root/sidecar

