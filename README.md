# `localtunnel` - expose yourself (in Go)

This repository contains a drop-in replacement of
[`localtunnel`](https://github.com/localtunnel/localtunnel) client CLI in Go. I
choose to rewrite it in Go to ensure static binary compilation, addressing
compatibility issues present in the original version (on Android notably).

## Getting started

First, build the go program as follow:

```
$ CGO_ENABLED=0 go build -o lt

# Compile for android
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -o lt
```

Then run it:

```
$ lt --help
Usage lt --port [num] <option>

Options:
      --allow-invalid-cert       Disable certificate checks for your local HTTPS server (ignore cert/key/ca options)
      --debug                    Show debug logs
      --help                     Show this help and exit
  -h, --host string              Upstream server providing forwarding (default "https://localtunnel.me")
      --local-ca string          Path to certificate authority file for self-signed certificates
      --local-cert string        Path to certificate PEM file for local HTTPS server
  -l, --local-host string        Tunnel traffic to this host instead of localhost, override Host header to this host (default "localhost")
      --local-https              Tunnel traffic to a local HTTPS server
      --local-key string         Path to certificate key file for local HTTPS server
  -m, --max-connections uint16   Max number of simultaneous connections (default 10)
  -o, --open                     Opens the tunnel URL in your browser
  -p, --port uint16              Internal HTTP server port
      --print-requests           No op, this is a compatibility flag
  -s, --subdomain string         Request this subdomain
      --version                  Show version number

$ lt --subdomain myapp --port 3000
your url is: https://myapp.loca.lt
...
```

## Contributing

If you want to contribute to `localtunnel` to add a feature or improve the code contact
me at [alexandre@negrel.dev](mailto:alexandre@negrel.dev), open an
[issue](https://github.com/negrel/localtunnel/issues) or make a
[pull request](https://github.com/negrel/localtunnel/pulls).

## :stars: Show your support

Please give a :star: if this project helped you!

[![buy me a coffee](https://github.com/negrel/.github/blob/master/.github/images/bmc-button.png?raw=true)](https://www.buymeacoffee.com/negrel)

## :scroll: License

MIT © [Alexandre Negrel](https://www.negrel.dev/)
