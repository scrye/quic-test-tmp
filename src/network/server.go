package network

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"

	quic "github.com/lucas-clemente/quic-go"
)

type Server struct {
	PacketConn net.PacketConn
	Out        io.Writer
}

func (s *Server) ListenAndServe() error {
	listener, err := quic.Listen(s.PacketConn, generateTLSConfig(), nil)
	if err != nil {
		return err
	}

	for {
		session, err := listener.Accept()
		if err != nil {
			return err
		}
		stream, err := session.AcceptStream()
		if err != nil {
			return err
		}
		_, err = io.Copy(stream, stream)
		if err != nil {
			return err
		}
		stream.Close()
	}
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
