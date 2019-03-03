FROM golang:latest AS builder
ARG GO111MODULE=on
WORKDIR /go/src/github.com/thoeni/godoc-create/
COPY . .
RUN go get ./... \
	&& go install . \
	&& go get github.com/thoeni/godocdash

FROM golang:latest
COPY --from=builder /go/bin/godocdash /usr/local/bin/
COPY --from=builder /go/bin/godoc-create /usr/local/bin/
COPY run.sh .

CMD ["sh run.sh"]