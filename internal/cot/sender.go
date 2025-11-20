package cot

import (
	"fmt"
	"net"
	"time"
)

const (
	MulticastAddr = "239.2.3.1:6969"
)

// Sender interface
type Sender interface {
	Send(cotXML []byte) error
	Close() error
}

// MulticastSender - UDP multicast
type MulticastSender struct {
	conn *net.UDPConn
}

func NewMulticastSender() (*MulticastSender, error) {
	addr, err := net.ResolveUDPAddr("udp", MulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve multicast addr: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial UDP: %w", err)
	}

	return &MulticastSender{conn: conn}, nil
}

func (s *MulticastSender) Send(cotXML []byte) error {
	_, err := s.conn.Write(cotXML)
	return err
}

func (s *MulticastSender) Close() error {
	return s.conn.Close()
}

// DirectSender - UDP direct
type DirectSender struct {
	conn *net.UDPConn
}

func NewDirectSender(targetIP string, port int) (*DirectSender, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetIP, port))
	if err != nil {
		return nil, fmt.Errorf("resolve target addr: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial UDP: %w", err)
	}

	return &DirectSender{conn: conn}, nil
}

func (s *DirectSender) Send(cotXML []byte) error {
	_, err := s.conn.Write(cotXML)
	return err
}

func (s *DirectSender) Close() error {
	return s.conn.Close()
}

// TCPSender - TCP connection to TAK Server
type TCPSender struct {
	conn net.Conn
	addr string
}

func NewTCPSender(serverIP string, port int) (*TCPSender, error) {
	addr := fmt.Sprintf("%s:%d", serverIP, port)

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to TAK server: %w", err)
	}

	return &TCPSender{conn: conn, addr: addr}, nil
}

func (s *TCPSender) Send(cotXML []byte) error {
	_, err := s.conn.Write(cotXML)
	if err != nil {
		// Try to reconnect once
		newConn, reconnectErr := net.DialTimeout("tcp", s.addr, 5*time.Second)
		if reconnectErr != nil {
			return fmt.Errorf("send failed and reconnect failed: %w", err)
		}
		s.conn.Close()
		s.conn = newConn
		_, err = s.conn.Write(cotXML)
	}
	return err
}

func (s *TCPSender) Close() error {
	return s.conn.Close()
}
