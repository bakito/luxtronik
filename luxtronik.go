package luxtronik

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CommandMode int32

const (
	DefaultPort                       = "8889"
	ParametersWrite       CommandMode = 3002
	ParametersRead        CommandMode = 3003
	CalculationsRead      CommandMode = 3004
	VisibilitiesRead      CommandMode = 3005
	SocketReadSizePeek                = 16
	SocketReadSizeInteger             = 4
	SocketReadSizeChar                = 1
)

// Locking is being used to ensure that only a single socket operation is
// performed at any point in time. This helps to avoid issues with the
// Luxtronik controller, which seems unstable otherwise.
var globalLock = &sync.Mutex{}

type Client interface {
	Connect() error
	Close() error
	ReadParameters(pm DataTypeMap) error
	WriteParameter(data ...int32) error
	ReadCalculations(pm DataTypeMap) error
	ReadVisibilities(pm DataTypeMap) error
}

type client struct {
	opts Options
	host string
	port string
	conn net.Conn
}

type Options struct {
	ConnCB      func(net.Conn) // gets called during connect to set conn specific params
	SafeMode    bool
	DialTimeout time.Duration
	Logger      *zap.Logger
}

func MustNewClient(hostPort string, opts Options) Client {
	c, err := MustNew(hostPort, opts)
	if err != nil {
		panic(err)
	}
	return c
}

func MustNew(hostPort string, opts Options) (Client, error) {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	if opts.DialTimeout < 1 {
		opts.DialTimeout = time.Minute
	}

	return &client{
		opts: opts,
		host: host,
		port: port,
	}, nil
}

func (c *client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *client) Connect() (err error) {
	if c.conn == nil {
		c.conn, err = net.DialTimeout("tcp", c.host+":"+c.port, c.opts.DialTimeout)
		if err != nil {
			return err
		}
		if c.opts.ConnCB != nil {
			c.opts.ConnCB(c.conn)
		}
	}

	return err
}

func (c *client) ReadParameters(pm DataTypeMap) error {
	return c.readFromHeatPump(pm, ParametersRead, 0)
}

func (c *client) WriteParameter(data ...int32) error {
	_, err := c.write(ParametersWrite, data...)
	return err
}

func (c *client) ReadCalculations(pm DataTypeMap) error {
	return c.readFromHeatPump(pm, CalculationsRead, 0)
}

func (c *client) ReadVisibilities(pm DataTypeMap) error {
	return c.readFromHeatPump(pm, VisibilitiesRead, 0)
}

func (c *client) readFromHeatPump(pm DataTypeMap, mode CommandMode, data ...int32) error {
	if len(data) < 1 {
		return fmt.Errorf("")
	}
	_, err := c.write(mode, data...)
	if err != nil {
		return fmt.Errorf("readFromHeatPump.netWrite to send %d failed: %w", data[0], err)
	}

	cmd, err := c.readInt32()
	if err != nil {
		return fmt.Errorf("readFromHeatPump.readInt32.cmd failed: %w", err)
	}

	if mode == CalculationsRead {
		var stat int32
		stat, err = c.readInt32()
		if err != nil {
			return fmt.Errorf("readFromHeatPump.readInt32.cmd failed: %w", err)
		}
		_ = stat
	}

	if cmd != int32(mode) {
		return fmt.Errorf("readFromHeatPump. received invalid command: %d want: %d", cmd, mode)
	}

	length, err := c.readInt32()
	if err != nil {
		return fmt.Errorf("readFromHeatPump.readInt32.length failed: %w", err)
	}

	rawValues := make([]int32, length)
	for i := int32(0); i < length; i++ {
		if mode == VisibilitiesRead {
			char, err := c.readChar()
			if err != nil {
				return fmt.Errorf("readFromHeatPump.readint32.paramID at index %d failed: %w", i, err)
			}
			rawValues[i] = int32(char) // 0 or 1
		} else {
			paramID, err := c.readInt32()
			if err != nil {
				return fmt.Errorf("readFromHeatPump.readint32.paramID at index %d failed: %w", i, err)
			}

			rawValues[i] = paramID
		}
	}

	return pm.SetRawValues(rawValues)
}

func (c *client) readInt32() (int32, error) {
	var buf [SocketReadSizeInteger]byte
	n, err := c.conn.Read(buf[:])
	if err != nil {
		return 0, err
	}

	return int32(binary.BigEndian.Uint32(buf[:n])), nil
}

func (c *client) readChar() (byte, error) {
	var buf [SocketReadSizeChar]byte
	n, err := c.conn.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n != SocketReadSizeChar {
		return 0, fmt.Errorf("read length %d is not equal to char size of %d", n, SocketReadSizeChar)
	}

	// res := binary.BigEndian.Uint32()
	return buf[0], nil
}

// readAndWrite
// Read and/or write value from and/or to heatpump.
// This method is essentially a wrapper for the _read() and _write()
// methods.
// Locking is being used to ensure that only a single socket operation is
// performed at any point in time. This helps to avoid issues with the
// Luxtronik controller, which seems unstable otherwise.
// If write is true, all parameters will be written to the heat pump
// prior to reading back in all data from the heat pump. If write is
// false, no data will be written, but all available data will be read
// from the heat pump.
// :param Parameters() parameters  Parameter dictionary to be written
//
//	to the heatpump before reading all available data
//	from the heatpump. At 'None' it is read only.
func (c *client) write(mode CommandMode, data ...int32) (int, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	var buf bytes.Buffer // refactor later
	payload := append([]int32{int32(mode)}, data...)
	if err := binary.Write(&buf, binary.BigEndian, payload); err != nil {
		return 0, fmt.Errorf("netWrite failed to encode: %#v with error: %w", payload, err)
	}

	return c.conn.Write(buf.Bytes())
}
