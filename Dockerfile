FROM golang:latest AS repo-fetcher
ARG GO111MODULE=on
WORKDIR /go/src/github.com/deliveroo/godoc/
COPY . .
RUN go get ./... \
	&& go build . \
	&& go get github.com/wuudjac/godocdash

FROM golang:latest AS deliveroo-repos
ARG github_token
ENV GITHUB_TOKEN=$github_token
COPY --from=repo-fetcher /go/src/github.com/deliveroo/godoc/godoc .
RUN ./godoc

FROM golang:latest
ENV GOPATH=/go
WORKDIR /go/src/github.com/deliveroo/
COPY --from=deliveroo-repos /tmp/deliveroo .
COPY --from=repo-fetcher /go/bin/godocdash .
EXPOSE 8080
CMD ["godoc", "-http=:8080"]