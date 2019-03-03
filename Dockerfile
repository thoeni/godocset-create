FROM golang:latest AS builder
ARG GO111MODULE=on
WORKDIR /go/src/github.com/thoeni/godocset-create/
COPY . .
RUN go get ./... \
	&& go install . \
	&& go get github.com/thoeni/godocdash

FROM golang:latest
ENV output=/tmp
ENV GOPHAT=/go
COPY --from=builder /go/bin/godocdash /usr/local/bin/
COPY --from=builder /go/bin/godocset-create /usr/local/bin/
COPY run.sh .

CMD ["/bin/bash","run.sh"]