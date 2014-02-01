package util

import (
	"errors"
	"net"
	"os"
)

func CreateSocket(socket string) (net.Listener, error) {
	c, err := net.Dial("unix", socket)

	if err == nil {
		c.Close()
		return nil, errors.New("socket in use")
	} else {
		os.Remove(socket)
	}

	l, err := net.Listen("unix", socket)

	if err != nil {
		return nil, err
	}

	os.Chmod(socket, 0777)

	return l, err
}
