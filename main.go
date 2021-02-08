package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func main() {
	var (
		mode       = flag.String("mode", "client", "[MTLS_PROXY_MODE] run mtls proxy as 'client' or 'server' mode")
		addr       = flag.String("addr", ":3000", "[MTLS_PROXY_ADDR] tcp address")
		caFile     = flag.String("ca", "", "[MTLS_PROXY_CA] ca cert file")
		certFile   = flag.String("cert", "", "[MTLS_PROXY_CERT] cert file")
		keyFile    = flag.String("key", "", "[MTLS_PROXY_KEY] key file")
		serverAddr = flag.String("server-addr", "", "[MTLS_PROXY_SERVER_ADDR] server address")
		serverName = flag.String("server-name", "", "[MTLS_PROXY_SERVER_NAME] server name")
	)

	flag.Parse()

	// override by env
	if x := os.Getenv("MTLS_PROXY_MODE"); x != "" {
		*mode = x
	}
	if x := os.Getenv("MTLS_PROXY_ADDR"); x != "" {
		*addr = x
	}
	if x := os.Getenv("MTLS_PROXY_CA"); x != "" {
		*caFile = x
	}
	if x := os.Getenv("MTLS_PROXY_CERT"); x != "" {
		*certFile = x
	}
	if x := os.Getenv("MTLS_PROXY_KEY"); x != "" {
		*keyFile = x
	}
	if x := os.Getenv("MTLS_PROXY_SERVER_ADDR"); x != "" {
		*serverAddr = x
	}
	if x := os.Getenv("MTLS_PROXY_SERVER_NAME"); x != "" {
		*serverName = x
	}

	if *serverAddr == "" {
		log.Fatalf("-server-addr required")
	}

	var cert tls.Certificate
	{
		var (
			certBytes []byte
			keyBytes  []byte
			err       error
		)
		if *certFile != "" {
			certBytes, err = ioutil.ReadFile(*certFile)
			if err != nil {
				log.Fatalf("can not load cert file; %v", err)
			}
		}
		if x := os.Getenv("MTLS_PROXY_CERT_BASE64"); x != "" {
			certBytes, err = base64.StdEncoding.DecodeString(x)
			if err != nil {
				log.Fatalf("can not load MTLS_PROXY_CERT_BASE64; %v", err)
			}
		}
		if *keyFile != "" {
			keyBytes, err = ioutil.ReadFile(*keyFile)
			if err != nil {
				log.Fatalf("can not load key file; %v", err)
			}
		}
		if x := os.Getenv("MTLS_PROXY_KEY_BASE64"); x != "" {
			keyBytes, err = base64.StdEncoding.DecodeString(x)
			if err != nil {
				log.Fatalf("can not load MTLS_PROXY_KEY_BASE64; %v", err)
			}
		}

		cert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			log.Fatalf("can not load x509 key pair; %v", err)
		}
	}

	ca := x509.NewCertPool()
	{
		var (
			caBytes []byte
			err     error
		)
		if *caFile != "" {
			caBytes, err = ioutil.ReadFile(*caFile)
			if err != nil {
				log.Fatalf("can not load ca file; %v", err)
			}
		}
		if x := os.Getenv("MTLS_PROXY_CA_BASE64"); x != "" {
			caBytes, err = base64.StdEncoding.DecodeString(x)
			if err != nil {
				log.Fatalf("can not load MTLS_PROXY_KEY_BASE64; %v", err)
			}
		}

		ok := ca.AppendCertsFromPEM(caBytes)
		if !ok {
			log.Fatalf("can not load ca")
		}
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    ca,
		ServerName:   *serverName,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}

	switch *mode {
	default:
		log.Fatalf("invalid mode")
	case "client", "server":
	}

	log.Printf("start mTLS Proxy %s mode on %s", *mode, *addr)
	switch *mode {
	case "client":
		dialer := tls.Dialer{
			Config: &tlsConfig,
		}

		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatalf("can not listen tcp; %v", err)
		}
		defer lis.Close()

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("accept connection error; %v", err)
				return
			}

			go func() {
				defer conn.Close()

				clientConn, err := dialer.Dial("tcp", *serverAddr)
				if err != nil {
					log.Printf("can not dial server; %v", err)
					return
				}
				defer clientConn.Close()

				tunnelConn(conn, clientConn)
			}()
		}
	case "server":
		lis, err := tls.Listen("tcp", *addr, &tlsConfig)
		if err != nil {
			log.Fatalf("can not listen tcp; %v", err)
		}
		defer lis.Close()

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("accept connection error; %v", err)
				return
			}

			go func() {
				defer conn.Close()

				clientConn, err := net.Dial("tcp", *serverAddr)
				if err != nil {
					log.Printf("can not dial server; %v", err)
					return
				}
				defer clientConn.Close()

				tunnelConn(conn, clientConn)
			}()
		}
	}
}

func tunnelConn(dst net.Conn, src net.Conn) {
	go func() {
		io.Copy(dst, src)
		dst.Close()
		src.Close()
	}()
	io.Copy(src, dst)
}
