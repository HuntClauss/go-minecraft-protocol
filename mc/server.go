package mc

import "net"

type Server struct {
	addr *net.TCPAddr
}

func NewServer(addr string) (Server, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	return Server{addr: a}, err
}

func (s Server) Connect() (*net.TCPConn, error) {
	conn, err := net.DialTCP("tcp", nil, s.addr)
	return conn, err
}
