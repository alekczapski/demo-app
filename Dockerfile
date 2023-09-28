FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG Version
ARG GitCommit

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/github.com/alekczapski/demo-app

COPY go.mod main.go main_test.go  .

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go test -v ./...

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags "-s -w -X main.release=${Version} -X main.commit=${GitCommit}" \
  -a -installsuffix cgo -o /usr/bin/demo-app .

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=https://github.com/alekczapski/demo-app

WORKDIR /
COPY --from=builder /usr/bin/demo-app /
USER nonroot:nonroot

CMD ["/demo-app"]
