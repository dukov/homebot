FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/src/homebot
COPY go.mod go.sum /usr/src/homebot/
RUN go mod download
COPY . /usr/src/homebot/
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
RUN env
RUN echo "OS $TARGETOS ARCH $TARGETARCH"
RUN go build -o /homebot
FROM alpine
COPY --from=build /homebot /homebot
