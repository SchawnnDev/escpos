package escpos

import "net"

type Printer interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

type networkPrinter struct {
	conn net.Conn
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

func NewNetworkPrinter(address string) (Printer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &networkPrinter{
		conn: conn,
	}, nil
}
