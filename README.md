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

You can find the docker image for this project available in Docker Hub: https://hub.docker.com/r/asanrom/turn-server

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

| Variable                      | Description                                                                                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `USERS`                       | List of users and passwords separated by commas. The user and the password must be separated by a colon (`:`).                                         |
| `AUTH_SECRET`                 | Secret to validate authentication tokens. Leave empty to disable auth tokens. Check the [authentication tokens documentation](#authentication-tokens). |
| `AUTH_CALLBACK_URL`           | URL of the authorization callback. Leave empty to disable it. Check the [authentication callback documentation](#authentication-callback)              |
| `AUTH_CALLBACK_AUTHORIZATION` | Value for the `Authorization` header when calling the authorization callback.                                                                          |

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

## Documentation

### Authentication tokens

In order to control the TURN server authentication with dynamic users, the following authentication token system is available:

- The `username` must follow the pattern: `turn/{TIMESTAMP}/{EXPIRATION}/{UID}`. The `TIMESTAMP` must be the token generation timestamp, in **UNIX time (Seconds)**. The `EXPIRATION` must be the expiration timestamp, also in **UNIX time (Seconds)**. The `UID` can be any string. It will be sent to the callback if configured.
- The `password` must be the **SHA-256** (SHA-2) of the UTF-8 bytes of the concatenation of the `username` and the `secret` (value of `AUTH_SECRET`), converted into **hexadecimal** and **lowercased**.

The application using the TURN server can generate these tokens as credentials for their users, controlling the duration of such credentials.

Here is an example in Go of the procedure of generation of the password:

```go
import (
  "crypto/sha256"
  "encoding/hex"
  "strings"
)

// Generates an authentication token, to be used
// as the password for the given username
//
// Parameters:
//   - username - The username
//   - secret - The secret shared between the TURN server and the application server
//
// Returns the password as string
func GenerateAuthPassword(username string, secret string) string {
	h := sha256.New()

	h.Write([]byte(username))
	h.Write([]byte(secret))

	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}
```

### Authentication callback

In order for your application to get a more fine-grained control of authentication, you can configure and URL for the TURN server in order to check for access.

Note: Authentication tokens must be used in order to also use the callback.

The procedure is the following:

- Every time the TURN server receives an authentication request, it will send a `GET` request to the URL provided by `AUTH_CALLBACK_URL`.
- To the URL, the following **query parameters** will be added:
  - `uid` = The `UID` part of the `username`
  - `ip` = The client IP address
- The value of `AUTH_CALLBACK_AUTHORIZATION` will be sent as the `Authorization` header, in order for the application to restrict access to the callback, so only the TURN server can use it.
- If the request to the callback returns an status code different from `200`, the authentication request is considered as failed, and the user will be denied of access to the TURN server.
- If the request to the callback returns an status code of `200`, the authentication process will continue, using a generated password as described in the [authentication tokens](#authentication-tokens) section.

### Using the TURN server

In order to use the TURN server in the browser, you can use the server by setting the URL, username and credential in the `iceServers` section of the `RTCPeerConnection` constructor options.

```js
const TURN_SERVER_HOST = "localhost";

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
