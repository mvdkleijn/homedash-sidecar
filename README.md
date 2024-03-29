# HomeDash sidecar

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mvdkleijn/homedash-sidecar?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvdkleijn/homedash-sidecar?style=for-the-badge)](https://goreportcard.com/report/github.com/mvdkleijn/homedash-sidecar) [![Liberapay patrons](https://img.shields.io/liberapay/patrons/mvdkleijn?style=for-the-badge)](https://liberapay.com/mvdkleijn/) [![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/O4O7H6C73)

This is the sidecar application that reports running containers to the HomeDash server.

Keep in mind that the sidecar application requires access to the Docker API.

See [homedash server](https://github.com/mvdkleijn/homedash) for more details.

## Usage

1) Start the binary release, providing the mandatory env variables; or
2) Start it as a container;
3) Configure some containers with appropriate labels;

See the repository root of [homedash server](https://github.com/mvdkleijn/homedash) for an example docker-compose.yml using the sidecar.

### Labels used

- `homedash.name=example`
- `homedash.url=http://my.example.com`
- `homedash.icon=example`
- `homedash.comment=Some random comment`

Everything except the `homedash.name` label is optional. Icons are retrieved from the SimpleIcons library by the server.

## Configuration

| Environment variable    | Description                                                 | Required | Default value                                   |
| ----------------------- | ----------------------------------------------------------- | -------- | ----------------------------------------------- |
| `HOMEDASH_SERVER`       | Full URL to HomeDash server. Example: http://localhost:8080 | Yes      | -                                               |
| `HOMEDASH_INTERVAL`     | Interval at which to check for apps. In minutes.            | No       | **10**                                          |
| `HOMEDASH_SIDECAR_UUID` | UUID to identify sidecar instance with.                     | No       | **(re)generated on every restart when missing** |

## Support

Supported Go versions, see: https://endoflife.date/go
Supported architectures: amd64, arm64

Source code and issues: https://github.com/mvdkleijn/homedash-sidecar

## Licensing

HomeDash sidecar is made available under the [MPL-2.0](https://choosealicense.com/licenses/mpl-2.0/)
license. The full details are available from the [LICENSE](/LICENSE) file.