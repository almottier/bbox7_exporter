FROM --platform=$BUILDPLATFORM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /bbox7_exporter .

FROM gcr.io/distroless/static-debian12
COPY --from=build /bbox7_exporter /bbox7_exporter
EXPOSE 9311
ENTRYPOINT ["/bbox7_exporter"]
