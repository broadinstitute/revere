ARG GO_VERSION='1.16'
ARG ALPINE_VERSION='3.14'

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build

ARG REVERE_BUILD_VERSION='development'

WORKDIR /build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOBIN=/bin
COPY . .
RUN go test ./... && go build -ldflags="-X 'main.BuildVersion=${REVERE_BUILD_VERSION}'" -o /bin/ .

FROM alpine:${ALPINE_VERSION} as runtime
COPY --from=build /bin/revere /bin/revere
ENTRYPOINT [ "/bin/revere" ]
