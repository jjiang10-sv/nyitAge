package transport

import (
	"testing"
	"time"
)

func TestTcpSession(t *testing.T) {
	go MainSession()
	time.Sleep(2 * time.Second)
	MainHandle()
}

func TestSsh(t *testing.T) {
	go MainSshServe()
	time.Sleep(2 * time.Second)
	MainSshClient()
}
