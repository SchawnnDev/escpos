package escpos

import (
	"net"
	"time"
)

type Printer interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

type networkPrinter struct {
	conn net.Conn
}

// PrinterOption defines a function that configures a network printer
type PrinterOption func(*networkPrinter) error

// WithDeadline sets both read and write deadlines
func WithDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetDeadline(t)
	}
}

// WithReadDeadline sets the deadline for future Read calls
func WithReadDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetReadDeadline(t)
	}
}

// WithWriteDeadline sets the deadline for future Write calls
func WithWriteDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetWriteDeadline(t)
	}
}

func NewNetworkPrinter(address string, opts ...PrinterOption) (Printer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	np := &networkPrinter{
		conn: conn,
	}

	for _, opt := range opts {
		if err = opt(np); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return np, nil
}

func (np *networkPrinter) Read(p []byte) (n int, err error) {
	return np.conn.Read(p)
}

func (np *networkPrinter) Write(p []byte) (n int, err error) {
	return np.conn.Write(p)
}

func (np *networkPrinter) Close() error {
	return np.conn.Close()
}
