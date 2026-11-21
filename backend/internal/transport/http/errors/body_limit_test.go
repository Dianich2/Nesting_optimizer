package httperror

import (
	"bufio"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestStage17BodyLimitIsClientError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: Handle})
	app.Post("/body", func(c fiber.Ctx) error { return c.SendStatus(204) })
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { defer close(done); _ = app.Listener(listener) }()
	t.Cleanup(func() { _ = app.Shutdown(); <-done })
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	// Advertised body is rejected before its bytes are read. No large allocation
	// or write/reset race from a streaming client is needed to test admission.
	_, err = fmt.Fprintf(conn, "POST /body HTTP/1.1\r\nHost: localhost\r\nContent-Type: application/json\r\nContent-Length: %d\r\nConnection: close\r\n\r\n", fiber.DefaultBodyLimit+1)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 413 {
		t.Fatalf("bounded oversized request returned%d want413", response.StatusCode)
	}
}
