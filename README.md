# HomeDash sidecar

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mvdkleijn/homedash-sidecar?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvdkleijn/homedash-sidecar?style=for-the-badge)](https://goreportcard.com/report/github.com/mvdkleijn/homedash-sidecar) [![Liberapay patrons](https://img.shields.io/liberapay/patrons/mvdkleijn?style=for-the-badge)](https://liberapay.com/mvdkleijn/) [![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/O4O7H6C73)

The sidecar application that reports running containers to the HomeDash server.
It supports both Docker and Docker Swarm. Easy configuration with only a single
required environment variable, the URL of the server.

Keep in mind that the sidecar application requires access to the Docker API.

See [HomeDash server](https://github.com/mvdkleijn/homedash) for more details.

## Usage

HomeDash sidecar can run as a standalone binary or as a container.

1) Configure some containers with the HomeDash labels;
2) Configure the mandatory environment variables;
3) Start the sidecar;

See the repository root of [HomeDash server](https://github.com/mvdkleijn/homedash)
for an example `compose.yml` using the sidecar.

### Labels used

- `homedash.name=example` (required)
- `homedash.url=http://my.example.com`
- `homedash.icon=example`
- `homedash.comment=Some random comment`

Everything except the `homedash.name` label is optional. Icons are provided by
the server and are essentially the same as the ones from Heimdall.

*) when using a custom label prefix, for example "myprefix", the label names change
   accordingly. In our example we'd get `myprefix.name` and `myprefix.url` for instance.

### Label placement

Labels can be placed in two spots. For plain old Docker, including Docker Compose,
the labels should be placed on the container level. For Docker Swarm, they can be
placed on either the container or service levels. If placed on both, the service
level lables will take precedence.

**Docker Compose example**

```yaml
services:
  myapp:
    image: XYZ
    labels:
      - homedash.name=My App
      - homedash.url=http://myapp.local/
      - homedash.icon=myapp
```

**Docker Swarm example**

```yaml
services:
  myapp:
    image: XYZ
    deploy:
      mode: replicated
      labels:
        - homedash.name=My App
        - homedash.url=http://myapp.local/
        - homedash.icon=myapp
```

## Configuration

| Environment variable    | Description                                             | Required | Default value                                   |
| ----------------------- | ------------------------------------------------------- | -------- | ----------------------------------------------- |
| `HOMEDASH_SERVER`       | Full URL to HomeDash server **without** trailing slash. | Yes      | -                                               |
| `HOMEDASH_INTERVAL`     | Interval at which to check for apps.                    | No       | "10m"                                           |
| `HOMEDASH_SIDECAR_UUID` | UUID to identify sidecar instance with.                 | No       | **(re)generated on every restart when missing** |
| `HOMEDASH_LABEL_PREFIX` | Allows setting a custom prefix for the labels.          | No       | "homedash"                                      |
| `HOMEDASH_LOG_LEVEL`    | Specifies logging level for sidecar application.        | No       | "INFO"                                          |

*) interval can be set as for exampple "10m", "5s" or "1h". See https://pkg.go.dev/time#ParseDuration for details.

## Support

Supported Go versions, see: https://endoflife.date/go
Supported architectures: amd64, arm64

Source code and issues: https://github.com/mvdkleijn/homedash-sidecar

## Licensing

HomeDash sidecar is made available under the [MPL-2.0](https://choosealicense.com/licenses/mpl-2.0/)
license. The full details are available from the [LICENSE](/LICENSE) file.