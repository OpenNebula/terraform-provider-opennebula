package testing

import (
	"fmt"
	"net"
	"net/rpc"
)

type OneServer struct {
	listener net.Listener
}

type One struct {
}

func (t *One) Template(args int, reply *int) error {
	reply = &args
	return nil
}

// Stop stops the server.
func (s *OneServer) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}

// URL returns the HTTP URL of the server.
func (s *OneServer) URL() string {
	if s.listener == nil {
		return ""
	}
	return "http://" + s.listener.Addr().String() + "/"
}

func NewServer(address string) (*OneServer, error) {
	handler := rpc.NewServer()
	tmp_one := new(One)
	handler.Register(tmp_one)

	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("listen(%q): %s\n", address, err)
		return nil, err
	}

	fmt.Printf("Server listening on %s\n", ln.Addr())
	go func() {
		for {
			cxn, err := ln.Accept()
			if err != nil {
				fmt.Printf("listen(%q): %s\n", address, err)
				return
			}
			fmt.Printf("Server ccepted connection to %s from %s\n", cxn.LocalAddr(), cxn.RemoteAddr())
			go handler.ServeConn(cxn)
		}
	}()
	server := OneServer{
		listener: ln,
	}
	return &server, nil
}
