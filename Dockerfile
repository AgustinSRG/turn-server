## Dockerfile ###

FROM golang:alpine

WORKDIR /root

ADD . /root/turn-server

WORKDIR /root/turn-server

# Compile

RUN go get .

RUN go build -ldflags="-s -w" .

# Prepare runner

FROM alpine as runner

# Add gcompat

RUN apk add gcompat

# Copy binaries

COPY --from=0 /root/turn-server/turn-server /bin/turn-server

# Expose ports

EXPOSE 3478/tcp
EXPOSE 3478/udp
EXPOSE 5349

EXPOSE 50000:55000/UDP

# Entry point

ENTRYPOINT ["/bin/turn-server"]