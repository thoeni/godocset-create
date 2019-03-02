## Deliveroo GoDoc

This repo contains what's needed to spin up a Docker container which runs a server with a copy of our Go repositories.

### Build:

#### Prerequisite:
Get a GitHub token from https://github.com/settings/tokens and **enable SSO**.

To build the Docker image, run this and replace the value with your authorised token:
```
docker build --build-arg github_token=745cc745cc745cc745cc745cc745cc745cc745cc -t godoc .
```

### Run:
```
docker run -p 8080:8080 godoc
```

Navigate to: http://localhost:8080/pkg/github.com/deliveroo/