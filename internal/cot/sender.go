package cot

import (
	"fmt"
	"net"
)

const (
	MulticastAddr = "239.2.3.1:6969" // Standard ATAK multicast
)

// Sender sends CoT messages
type Sender interface {
	Send(cotXML []byte) error
	Close() error
}

// MulticastSender sends via UDP multicast
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
