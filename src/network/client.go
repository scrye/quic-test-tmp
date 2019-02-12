package network

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	quic "github.com/lucas-clemente/quic-go"
)

type Client struct {
	PacketConn    net.PacketConn
	RemoteAddress string

	Out io.Writer
}

func (c *Client) Run(r io.Reader, w io.Writer) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", c.RemoteAddress)
	if err != nil {
		return fmt.Errorf("udp address parse error, err = %v", err)
	}

	session, err := quic.Dial(c.PacketConn, udpAddr, c.RemoteAddress, &tls.Config{InsecureSkipVerify: true}, &quic.Config{})
	if err != nil {
		return fmt.Errorf("quic dial error, err = %v", err)
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		return fmt.Errorf("stream open error, err = %v", err)
	}
	defer stream.Close()

	go c.Sender(stream, r)
	return c.Receiver(stream, w)
}

func (c *Client) Sender(stream quic.Stream, r io.Reader) error {
	_, err := io.Copy(stream, r)
	if err != nil {
		return err
	}
	return stream.Close()

}

func (c *Client) Receiver(stream quic.Stream, w io.Writer) error {
	_, err := io.Copy(w, stream)
	if err != nil {
		return err
	}
	return nil
}
