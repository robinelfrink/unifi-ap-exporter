FROM --platform=$BUILDPLATFORM golang:1.24.5 AS BUILD
WORKDIR /app
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=development
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s -X 'main.Version=${VERSION}'" .

FROM scratch
WORKDIR /app
COPY --from=BUILD /app/unifi-ap-exporter /app

CMD [ "/app/unifi-ap-exporter" ]
