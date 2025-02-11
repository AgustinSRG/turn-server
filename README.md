# TURN server

This a TURN server for WebRTC applications, built using the [Pion TURN](https://github.com/pion/turn) toolkit.

## Compilation

[Go](https://go.dev/doc/install) is required to compile the project.

In order to install dependencies, type:

```
go get .
```

To compile the code type:

```sh
go build .
```

The build command will create a binary in the current directory, called `turn-server`, or `turn-server.exe` if you are using Windows.

## Docker

You can find the docker image for this project available in Docker Hub: https://hub.docker.com/r/asanrom/hls-websocket-cdn

To pull it type:

```sh
docker pull asanrom/turn-server
```

Example compose file:

```yaml
version: "3.7"

services:
  cdn_server:
    image: asanrom/turn-server
    ports:
      - "3478:3478/tcp"
      - "3478:3478/udp"
      - "5349:5349"
      - "50000-55000:50000-55000/udp"
    environment:
      # Configure it using env vars:
      - REALM=example.com
      - PUBLIC_IP=10.0.0.1
      - MIN_RELAY_PORT=50000
      - MAX_RELAY_PORT=55000
      - USERS=user:password,user2:password2
```

## Configuration

You can configure the server using environment variables.

### TURN server configuration

| Variable    | Description                                                                              |
| ----------- | ---------------------------------------------------------------------------------------- |
| `REALM`     | Realm for the TURN server. Set your domain. Example: `example.com`                       |
| `PUBLIC_IP` | External IP address of the server. If you leave it blank, it will try to auto detect it. |

### RTP relay port range

| Variable         | Description                                         |
| ---------------- | --------------------------------------------------- |
| `MIN_RELAY_PORT` | Start of the RTP relay port range. Default: `50000` |
| `MAX_RELAY_PORT` | End of the RTP relay port range. Default: `55000`   |

Note: The ports must be opened through the firewall under the UDP protocol.

### Authentication

| Variable | Description                                                                                                    |
| -------- | -------------------------------------------------------------------------------------------------------------- |
| `USERS`  | List of users and passwords separated by commas. The user and the password must be separated by a colon (`:`). |

### UDP Listener

| Variable           | Description                                                                     |
| ------------------ | ------------------------------------------------------------------------------- |
| `UDP_ENABLED`      | Can be `YES` or `NO`. Set it to `YES` in order to enable the UDP listener.      |
| `UDP_PORT`         | The port number for the UDP listener (`3478` by default)                        |
| `UDP_BIND_ADDRESS` | The bind address UDP listener (Leave empty to listen on all network interfaces) |

### TCP Listener

| Variable           | Description                                                                     |
| ------------------ | ------------------------------------------------------------------------------- |
| `TCP_ENABLED`      | Can be `YES` or `NO`. Set it to `YES` in order to enable the TCP listener.      |
| `TCP_PORT`         | The port number for the TCP listener (`3478` by default)                        |
| `TCP_BIND_ADDRESS` | The bind address TCP listener (Leave empty to listen on all network interfaces) |

### TLS Listener

| Variable                   | Description                                                                         |
| -------------------------- | ----------------------------------------------------------------------------------- |
| `TLS_ENABLED`              | Can be `YES` or `NO`. Set it to `YES` in order to enable the TLS listener.          |
| `TLS_PORT`                 | The port number for the TLS listener (`5349` by default)                            |
| `TLS_BIND_ADDRESS`         | The bind address TLS listener (Leave empty to listen on all network interfaces)     |
| `TLS_CERTIFICATE`          | Path to the X.509 certificate for TLS                                               |
| `TLS_PRIVATE_KEY`          | Path to the private key for TLS                                                     |
| `TLS_CHECK_RELOAD_SECONDS` | Number of seconds to check for changes in the certificate or key (for auto renewal) |

### Log configuration

| Variable      | Description                                                                                         |
| ------------- | --------------------------------------------------------------------------------------------------- |
| `LOG_ERROR`   | Can be `YES` or `NO`. Default: `YES`. Set it to `YES` in order to enable logging `ERROR` messages   |
| `LOG_WARNING` | Can be `YES` or `NO`. Default: `YES`. Set it to `YES` in order to enable logging `WARNING` messages |
| `LOG_INFO`    | Can be `YES` or `NO`. Default: `YES`. Set it to `YES` in order to enable logging `INFO` messages    |
| `LOG_DEBUG`   | Can be `YES` or `NO`. Default: `NO`. Set it to `YES` in order to enable logging `DEBUG` messages    |
| `LOG_TRACE`   | Can be `YES` or `NO`. Default: `NO`. Set it to `YES` in order to enable logging `TRACE` messages    |

## Using the TURN server

Assuming you set an user with name `user` and password `password`, you can user the server by setting the URL, username and credential in the `iceServers` section of the `RTCPeerConnection` constructor options.

```js
const TURN_SERVER_HOST = "localhost"

const iceConfiguration = {
  iceServers: [
    {
      urls: [
        // Add the TURN server URLs for all the available transports
        "turn:" + TURN_SERVER_HOST + ":3478", // UDP
        "turn:" + TURN_SERVER_HOST + ":3478?transport=tcp", // TCP
        "turns:" + TURN_SERVER_HOST + ":5349?transport=tcp", // TLS
      ],
      // Use the credentials for the server
      username: "user",
      credential: "password",
    },
  ],
};

const peerConnection = new RTCPeerConnection(iceConfiguration);
```

You can also use the tool to test your TURN server: https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/
