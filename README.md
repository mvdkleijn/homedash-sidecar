# HomeDash sidecar

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

| Environment variable | Description | Required | Default value |
| -------------------- | ----------- | -------- | ------------- |
| `HOMEDASH_SERVER`    | Full URL to HomeDash server. Example: http://localhost:8080 | Yes | - |
| `HOMEDASH_INTERVAL`  | Interval at which to check for apps. | No | **10m** |

## Support

Supported Go versions, see: https://endoflife.date/go
Supported architectures: amd64, arm64

Source code and issues: https://github.com/mvdkleijn/homedash-sidecar

## Licensing

HomeDash sidecar is made available under the [MPL-2.0](https://choosealicense.com/licenses/mpl-2.0/)
license. The full details are available from the [LICENSE](/LICENSE) file.