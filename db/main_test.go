package db_test

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	cmd := exec.Command("redis-server")
	cmd.Start()

	time.Sleep(time.Second * 2)

	res := m.Run()

	cmd.Process.Kill()

	os.Exit(res)
}
