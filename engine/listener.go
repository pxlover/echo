package engine

import (
	"net"
	"strings"
	"time"

	reuseport "github.com/admpub/go-reuseport"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return tc, err
	}
	err = tc.SetKeepAlive(true)
	if err != nil {
		return tc, err
	}
	err = tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, err
}

func NewListener(address string, reuse bool) (net.Listener, error) {
	scheme := "tcp"
	delim := "://"
	if pos := strings.Index(address, delim); pos > 0 {
		scheme = address[0:pos]
		address = address[pos+len(delim):]
	}
	var (
		l   net.Listener
		err error
	)
	if reuse {
		l, err = reuseport.Listen(scheme, address)
	} else {
		l, err = net.Listen(scheme, address)
	}
	if err != nil {
		return nil, err
	}
	switch listener := l.(type) {
	case *net.TCPListener:
		return &tcpKeepAliveListener{listener}, nil
	case *net.UnixListener:
		return listener, nil
	default:
		return l, nil
	}
}
