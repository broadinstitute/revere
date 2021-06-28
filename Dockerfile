ARG GO_VERSION='1.16'
ARG ALPINE_VERSION='3.13'

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build
WORKDIR /build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOBIN=/bin
COPY . .
RUN go test ./... && go build -o /bin/ .

FROM alpine:${ALPINE_VERSION} as runtime
ENV APP_NAME=revere
COPY --from=build /bin/${APP_NAME} /bin/${APP_NAME}
ENTRYPOINT [ "sh", "-c", "/bin/${APP_NAME}" ]