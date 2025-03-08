# TorBlockRedirect

TorBlockRedirect is a [Traefik](https://traefik.io) plugin that can block requests originating from the Tor network. The publicly available list of Tor exit nodes (`https://check.torproject.org/exit-addresses`) is fetched regularly to identify requests to block. The plugin also supports redirecting requests from Tor users to an onion site if the `OnionHostname` configuration is set. Additionally, the plugin now supports both IPv4 and IPv6 address blocking.

## Configuration

Requirements:
- `Traefik >= v2.5.5`

### Static

For each plugin, the Traefik static configuration must define the module name (as is usual for Go packages).

The following declaration (given here in YAML) defines a plugin:

```yaml
# Static configuration
pilot:
  token: xxxxx

experimental:
  plugins:
    torblockredirect:
      moduleName: github.com/PaulLeRoux142/TorBlockRedirect
      version: v0.1.2
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the http.middlewares section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - my-middleware

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    my-middleware:
      plugin:
        torblockredirect:
          enabled: true # default 'true'
#          UpdateIntervalSeconds: 3600 # default '3600'
#          OnionHostname: "YOUR_ONION_DOMAIN.onion" # default '' - block tor users if not set
#          AddressListURL: "https://www.dan.me.uk/torlist/?exit" # default 'https://check.torproject.org/exit-addresses'
#          ForwardedHeadersCustomName: "CF_CONNECTING_IP" # default 'X-Forwarded-For'
```

### Local Mode

Traefik also offers a developer mode that can be used for temporary testing or offline usage of plugins not hosted on GitHub. To use a plugin in local mode, the Traefik static configuration must define the module name (as is usual for Go packages) and a path to a [Go workspace](https://golang.org/doc/gopath_code.html#Workspaces), which can be the local GOPATH or any directory.

The plugins must be placed in `./plugins-local` directory, which should be next to the Traefik binary.
The source code of the plugin should be organized as follows:

```
./plugins-local/
    └── src
        └── github.com
            └── PaulLeRoux142
                └── TorBlockRedirect
                    ├── .golangci.toml
                    ├── .traefik.yml
                    ├── go.mod
                    ├── go.sum
                    ├── LICENSE
                    ├── Makefile
                    ├── netaddr.go
                    ├── README.md
                    ├── torblockredirect_test.go
                    ├── torblockredirect.go
                    └── examples
                        └── docker-compose.yml
```

```yaml
# Static configuration
pilot:
  token: xxxxx

experimental:
  localPlugins:
    example:
      moduleName: github.com/PaulLeRoux142/TorBlockRedirect
```

(In the above example, the `TorBlockRedirect` plugin will be loaded from the path `./plugins-local/src/github.com/PaulLeRoux142/TorBlockRedirect`)

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - my-middleware

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    my-middleware:
      plugin:
        torblockredirect:
          enabled: true # default 'true'
#          UpdateIntervalSeconds: 3600 # default '3600'
#          OnionHostname: "YOUR_ONION_DOMAIN.onion" # default '' - block tor users if not set
#          AddressListURL: "https://www.dan.me.uk/torlist/?exit" # default 'https://check.torproject.org/exit-addresses'
#          ForwardedHeadersCustomName: "CF_CONNECTING_IP" # default 'X-Forwarded-For'
```

## Features

- **Block Tor Requests**: The plugin blocks incoming requests from the Tor network.
- **Redirect to Onion Site**: If the `OnionHostname` configuration is set, requests from Tor users will be redirected to the specified `.onion` domain.
- **IPv6 Support**: The plugin now supports both IPv4 and IPv6 addresses when identifying and blocking Tor exit node IPs.

### Examples

You can also see a working example docker-compose.yml in the examples directory, which loads the plugin in local mode.
```
./examples/docker-compose.yml
```

### Plugin Configuration

The plugin supports the following configuration options:

- `enabled`: Enables or disables the plugin. Default is `true`.
- `OnionHostname`: If set, requests from Tor users will be redirected to this `.onion` site.
- `UpdateIntervalSeconds`: Interval in seconds for updating the list of blocked Tor exit nodes. Default is `3600` (1 hour).
- `AddressListURL`: URL to fetch the list of Tor exit nodes. Default is `https://check.torproject.org/exit-addresses`.
- `ForwardedHeadersCustomName`: Header name for the forwarded client IP address. Default is `X-Forwarded-For`.

# Example configuration for OnionHostname:
```yaml
torblockredirect:
  enabled: true
  OnionHostname: "youroniondomain.onion"
```