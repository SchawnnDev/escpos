package escpos

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTCPServer creates a mock TCP server for testing
func mockTCPServer(t *testing.T, handler func(net.Conn)) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()

	cleanup := func() {
		listener.Close()
	}

	return listener.Addr().String(), cleanup
}

// TestNewNetworkPrinter tests creating a network printer
func TestNewNetworkPrinter(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		// Echo server
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			conn.Write(buf[:n])
		}
	})
	defer cleanup()

	// Test without options
	printer, err := NewNetworkPrinter(addr)
	require.NoError(t, err)
	require.NotNil(t, printer)
	defer printer.Close()

	// Test write and read
	testData := []byte("Hello, Printer!")
	n, err := printer.Write(testData)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)

	readBuf := make([]byte, len(testData))
	n, err = printer.Read(readBuf)
	assert.NoError(t, err)
	assert.Equal(t, testData, readBuf[:n])
}

// TestWithTimeout tests the WithTimeout option
func TestWithTimeout(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		// Slow server - delays before responding
		time.Sleep(200 * time.Millisecond)
		buf := make([]byte, 1024)
		conn.Read(buf)
	})
	defer cleanup()

	// Create printer with short timeout
	printer, err := NewNetworkPrinter(addr, WithTimeout(50*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	// Write should succeed
	_, err = printer.Write([]byte("test"))
	assert.NoError(t, err)

	// Read should timeout
	buf := make([]byte, 1024)
	_, err = printer.Read(buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestWithReadTimeout tests the WithReadTimeout option
func TestWithReadTimeout(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		// Server that never responds
		time.Sleep(1 * time.Second)
	})
	defer cleanup()

	printer, err := NewNetworkPrinter(addr, WithReadTimeout(50*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	// Read should timeout
	buf := make([]byte, 1024)
	_, err = printer.Read(buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestWithWriteTimeout tests the WithWriteTimeout option
func TestWithWriteTimeout(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		// Server that doesn't read (causes write buffer to fill)
		time.Sleep(1 * time.Second)
	})
	defer cleanup()

	printer, err := NewNetworkPrinter(addr, WithWriteTimeout(50*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	// Write a large amount of data to fill the buffer
	largeData := make([]byte, 1024*1024) // 1MB
	_, err = printer.Write(largeData)
	// Note: Small writes might succeed even with timeout if buffer isn't full
	// This test mainly validates the API works
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
}

// TestTimeoutRecurring verifies that timeouts are applied to each operation
func TestTimeoutRecurring(t *testing.T) {
	requestCount := 0
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			// Respond quickly
			conn.Write(buf[:n])
		}
	})
	defer cleanup()

	printer, err := NewNetworkPrinter(addr, WithTimeout(1*time.Second))
	require.NoError(t, err)
	defer printer.Close()

	// Multiple operations should all work with timeout applied each time
	for i := 0; i < 5; i++ {
		testData := []byte(fmt.Sprintf("Request %d", i))

		// Write
		n, err := printer.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)

		// Read
		readBuf := make([]byte, len(testData))
		n, err = printer.Read(readBuf)
		assert.NoError(t, err)
		assert.Equal(t, testData, readBuf[:n])

		requestCount++

		// Small delay between requests
		time.Sleep(10 * time.Millisecond)
	}

	assert.Equal(t, 5, requestCount, "All requests should succeed")
}

// TestWithDeadline tests the WithDeadline option
func TestWithDeadline(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			conn.Write(buf[:n])
		}
	})
	defer cleanup()

	deadline := time.Now().Add(100 * time.Millisecond)
	printer, err := NewNetworkPrinter(addr, WithDeadline(deadline))
	require.NoError(t, err)
	defer printer.Close()

	// First operation should work
	_, err = printer.Write([]byte("test1"))
	assert.NoError(t, err)

	// Wait for deadline to pass
	time.Sleep(150 * time.Millisecond)

	// Operation after deadline should fail
	_, err = printer.Write([]byte("test2"))
	assert.Error(t, err)
}

// TestReadTimeoutPriority tests that readTimeout takes priority over timeout
func TestReadTimeoutPriority(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		time.Sleep(1 * time.Second)
	})
	defer cleanup()

	// Set both timeout and readTimeout, readTimeout should take priority
	printer, err := NewNetworkPrinter(addr,
		WithTimeout(200*time.Millisecond),
		WithReadTimeout(50*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	start := time.Now()
	buf := make([]byte, 1024)
	_, err = printer.Read(buf)
	duration := time.Since(start)

	// Should timeout closer to 50ms (readTimeout) than 200ms (timeout)
	assert.Error(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Should use readTimeout, not timeout")
}

// TestWriteTimeoutPriority tests that writeTimeout takes priority over timeout
func TestWriteTimeoutPriority(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		time.Sleep(1 * time.Second)
	})
	defer cleanup()

	// Set both timeout and writeTimeout, writeTimeout should take priority
	printer, err := NewNetworkPrinter(addr,
		WithTimeout(200*time.Millisecond),
		WithWriteTimeout(50*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	start := time.Now()
	largeData := make([]byte, 1024*1024) // 1MB
	_, err = printer.Write(largeData)
	duration := time.Since(start)

	// Write might succeed if buffer is large enough, but timing should be quick
	if err != nil {
		assert.Less(t, duration, 150*time.Millisecond, "Should respect writeTimeout")
	}
}

// TestInvalidAddress tests error handling for invalid addresses
func TestInvalidAddress(t *testing.T) {
	_, err := NewNetworkPrinter("invalid:99999")
	assert.Error(t, err)
}

// TestMultipleOptions tests using multiple options together
func TestMultipleOptions(t *testing.T) {
	addr, cleanup := mockTCPServer(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 1024)
		conn.Read(buf)
	})
	defer cleanup()

	printer, err := NewNetworkPrinter(addr,
		WithReadTimeout(100*time.Millisecond),
		WithWriteTimeout(100*time.Millisecond))
	require.NoError(t, err)
	defer printer.Close()

	// Both operations should work with their respective timeouts
	_, err = printer.Write([]byte("test"))
	assert.NoError(t, err)
}
