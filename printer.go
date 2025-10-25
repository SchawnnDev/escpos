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
	conn           net.Conn
	timeout        time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	connectTimeout time.Duration
}

// PrinterOption defines a function that configures a network printer
type PrinterOption func(*networkPrinter) error

// WithTimeout sets a timeout duration for all Read and Write operations
func WithTimeout(d time.Duration) PrinterOption {
	return func(np *networkPrinter) error {
		np.timeout = d
		return nil
	}
}

// WithConnectTimeout sets the timeout duration for the initial connection establishment
func WithConnectTimeout(d time.Duration) PrinterOption {
	return func(np *networkPrinter) error {
		np.connectTimeout = d
		return nil
	}
}

// WithReadTimeout sets the timeout duration for Read operations
func WithReadTimeout(d time.Duration) PrinterOption {
	return func(np *networkPrinter) error {
		np.readTimeout = d
		return nil
	}
}

// WithWriteTimeout sets the timeout duration for Write operations
func WithWriteTimeout(d time.Duration) PrinterOption {
	return func(np *networkPrinter) error {
		np.writeTimeout = d
		return nil
	}
}

// WithDeadline sets both read and write deadlines to an absolute time
// Note: This sets a one-time deadline. For recurring timeouts, use WithTimeout instead.
func WithDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetDeadline(t)
	}
}

// WithReadDeadline sets the deadline for future Read calls to an absolute time
// Note: This sets a one-time deadline. For recurring timeouts, use WithReadTimeout instead.
func WithReadDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetReadDeadline(t)
	}
}

// WithWriteDeadline sets the deadline for future Write calls to an absolute time
// Note: This sets a one-time deadline. For recurring timeouts, use WithWriteTimeout instead.
func WithWriteDeadline(t time.Time) PrinterOption {
	return func(np *networkPrinter) error {
		return np.conn.SetWriteDeadline(t)
	}
}

func NewNetworkPrinter(address string, opts ...PrinterOption) (Printer, error) {
	np := &networkPrinter{}

	// Apply options first to get the connectTimeout
	for _, opt := range opts {
		if err := opt(np); err != nil {
			return nil, err
		}
	}

	// Use net.Dialer with timeout if connectTimeout is set
	var conn net.Conn
	var err error
	if np.connectTimeout > 0 {
		d := net.Dialer{Timeout: np.connectTimeout}
		conn, err = d.Dial("tcp", address)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	np.conn = conn
	return np, nil
}

func (np *networkPrinter) Read(p []byte) (n int, err error) {
	// Set read deadline before each read operation
	if np.readTimeout > 0 {
		if err = np.conn.SetReadDeadline(time.Now().Add(np.readTimeout)); err != nil {
			return 0, err
		}
	} else if np.timeout > 0 {
		if err = np.conn.SetReadDeadline(time.Now().Add(np.timeout)); err != nil {
			return 0, err
		}
	}
	return np.conn.Read(p)
}

func (np *networkPrinter) Write(p []byte) (n int, err error) {
	// Set write deadline before each write operation
	if np.writeTimeout > 0 {
		if err = np.conn.SetWriteDeadline(time.Now().Add(np.writeTimeout)); err != nil {
			return 0, err
		}
	} else if np.timeout > 0 {
		if err = np.conn.SetWriteDeadline(time.Now().Add(np.timeout)); err != nil {
			return 0, err
		}
	}
	return np.conn.Write(p)
}

func (np *networkPrinter) Close() error {
	return np.conn.Close()
}
