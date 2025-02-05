package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	lt "github.com/jonasfj/go-localtunnel"
	"github.com/spf13/pflag"
)

func main() {
	port := pflag.Uint16P("port", "p", 0, "Internal HTTP server port")
	host := pflag.StringP("host", "h", "https://localtunnel.me", "Upstream server providing forwarding")
	subdomain := pflag.StringP("subdomain", "s", "", "Request this subdomain")
	localHost := pflag.StringP("local-host", "l", "localhost", "Tunnel traffic to this host instead of localhost, override Host header to this host")
	localHttps := pflag.Bool("local-https", false, "Tunnel traffic to a local HTTPS server")
	localCert := pflag.String("local-cert", "", "Path to certificate PEM file for local HTTPS server")
	localKey := pflag.String("local-key", "", "Path to certificate key file for local HTTPS server")
	localCa := pflag.String("local-ca", "", "Path to certificate authority file for self-signed certificates")
	allowInvalidCert := pflag.Bool("allow-invalid-cert", false, "Disable certificate checks for your local HTTPS server (ignore cert/key/ca options)")
	open := pflag.BoolP("open", "o", false, "Opens the tunnel URL in your browser")
	printRequests := pflag.Bool("print-requests", false, "No op, this is a compatibility flag")
	help := pflag.Bool("help", false, "Show this help and exit")
	version := pflag.Bool("version", false, "Show version number")

	// Extra flags.
	debug := pflag.Bool("debug", false, "Show debug logs")
	maxConnections := pflag.Uint16P("max-connections", "m", 10, "Max number of simultaneous connections")
	pflag.Parse()

	// Compatibility flags.
	_ = printRequests

	logger := log.New(os.Stderr, "[DEBUG] ", 0)
	if !*debug {
		logger.SetOutput(io.Discard)
	}

	var tlsConfig *tls.Config
	if *localHttps {
		tlsConfig = &tls.Config{
			ServerName: *localHost,
		}

		if *allowInvalidCert {
			tlsConfig.InsecureSkipVerify = true
		}

		if *localCert == "" || *localKey == "" {
			log.Fatal("invalid args: --local-key or --local-cert is undefined")
		} else {
			cert, err := tls.LoadX509KeyPair(*localCert, *localKey)
			if err != nil {
				log.Fatal("failed to load X509 key pair:", err)
			}

			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		if *localCa != "" {
			caCert, err := os.ReadFile(*localCa)
			if err != nil {
				log.Fatal("failed to read CA file:", err)
			}

			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				log.Fatal("failed to append PEM encoded certificates to the pool: no certificate found")
			}

			tlsConfig.ClientCAs = caCertPool
		}
	}

	if *help {
		printUsage()
		return
	}

	if *version {
		println("0.1.0")
		return
	}

	if *port == 0 {
		println("missing required argument: port")
		printUsage()
		os.Exit(1)
	}

	// Connect to localtunnel server.
	listener, err := lt.Listen(lt.Options{
		Subdomain:      *subdomain,
		BaseURL:        *host,
		MaxConnections: int(*maxConnections),
		Log:            logger,
	})
	if err != nil {
		log.Fatal("failed to initialize local tunnel", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Fatal("failed to close local tunnel connection")
		}
	}()

	fmt.Println("your url is:", listener.URL())

	// Open URL.
	if *open {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", listener.URL())
		case "darwin":
			cmd = exec.Command("open", listener.URL())
		case "windows":
			cmd = exec.Command("explorer.exe", listener.URL())
		}

		if cmd != nil {
			err := cmd.Start()
			if err != nil {
				log.Printf("failed to open URL")
			}
		} else {
			log.Printf("opening URL on %v is not supported\n", runtime.GOOS)
		}
	}

	// Handle SIGINT.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig, "received, exiting...")
		cancel()
		if err := listener.Close(); err != nil {
			log.Fatal("failed to stop localtunnel listener:", err)
		}
	}()

	// Start forwarding connections.
	upstreamAddr := fmt.Sprintf("%v:%v", *localHost, *port)
loop:
	for {
		downStream, err := listener.Accept()
		if err != nil && err.Error() == "listener was closed" {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
		}
		if err != nil {
			fmt.Printf("failed to accept connection: %v\n", err.Error())
			continue
		}

		log.Println("new connection", downStream.LocalAddr())
		go forward(ctx, logger, downStream, upstreamAddr, tlsConfig)
	}
}

func printUsage() {
	println("Usage lt --port [num] <option>")
	println("")
	println("Options:")
	pflag.PrintDefaults()
}

func forward(ctx context.Context, logger *log.Logger, downstream net.Conn, upstreamAddr string, tlsConfig *tls.Config) {
	var dialer net.Dialer
	tcpUpstream, err := dialer.DialContext(ctx, "tcp", upstreamAddr)
	if errors.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		log.Fatal("failed to connect to upstream service:", err)
	}
	tcpUpstream.(*net.TCPConn).SetKeepAlive(true)

	var upstream net.Conn = tcpUpstream
	if tlsConfig != nil {
		upstream = tls.Client(tcpUpstream, tlsConfig)
	}

	go func() {
		_, err := io.Copy(downstream, upstream)
		if err != nil {
			logger.Println("stream error:", err.Error())
		}
	}()
	go func() {
		_, err := io.Copy(upstream, downstream)
		if err != nil {
			logger.Println("stream error:", err.Error())
		}
	}()
}
