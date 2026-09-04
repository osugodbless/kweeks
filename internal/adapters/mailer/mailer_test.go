package mailer

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeSMTPServer accepts a connection, consumes the SMTP dialogue, and returns
// the raw message lines that were sent.
func fakeSMTPServer(t *testing.T) (addr string, got func() string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu struct {
		lines []string
		done  chan struct{}
	}
	mu.done = make(chan struct{})
	go func() {
		defer close(mu.done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		w := bufio.NewWriter(conn)
		r := bufio.NewReader(conn)
		write := func(s string) { _, _ = w.WriteString(s); _ = w.Flush() }
		write("220 fake ESMTP\r\n")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			mu.lines = append(mu.lines, line)
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				write("250-fake\r\n250 AUTH PLAIN LOGIN\r\n")
			case strings.HasPrefix(upper, "AUTH"):
				write("235 ok\r\n")
			case strings.HasPrefix(upper, "MAIL FROM"):
				write("250 ok\r\n")
			case strings.HasPrefix(upper, "RCPT TO"):
				write("250 ok\r\n")
			case strings.HasPrefix(upper, "DATA"):
				write("354 go ahead\r\n")
				// read until the lone dot
				for {
					bodyLine, err := r.ReadString('\n')
					if err != nil {
						return
					}
					bodyLine = strings.TrimRight(bodyLine, "\r\n")
					mu.lines = append(mu.lines, bodyLine)
					if bodyLine == "." {
						write("250 queued\r\n")
						break
					}
				}
			case strings.HasPrefix(upper, "QUIT"):
				write("221 bye\r\n")
				return
			default:
				write("250 ok\r\n")
			}
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return ln.Addr().String(), func() string {
		select {
		case <-mu.done:
		case <-time.After(3 * time.Second):
		}
		return strings.Join(mu.lines, "\n")
	}
}

func TestSendRedemptionEmailOverSMTP(t *testing.T) {
	addr, got := fakeSMTPServer(t)
	m := New(hostOf(addr), portOf(addr), "sender@gmail.com", "apppw", "sender@gmail.com", "", nil)

	if err := m.SendRedemptionEmail(context.Background(), "zainab@x.com", "KWEKS-ABC", "15000"); err != nil {
		t.Fatalf("send: %v", err)
	}
	out := got()
	for _, want := range []string{"MAIL FROM:<sender@gmail.com>", "RCPT TO:<zainab@x.com>", "Claim code: KWEKS-ABC"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestMailToOverrideRewritesRecipient(t *testing.T) {
	addr, got := fakeSMTPServer(t)
	m := New(hostOf(addr), portOf(addr), "sender@gmail.com", "apppw", "sender@gmail.com", "operator@example.com", nil)

	if err := m.SendRedemptionEmail(context.Background(), "zainab@x.com", "KWEKS-ABC", "15000"); err != nil {
		t.Fatalf("send: %v", err)
	}
	out := got()
	if strings.Contains(out, "zainab@x.com") {
		t.Fatalf("override failed; recipient still the winner:\n%s", out)
	}
	if !strings.Contains(out, "operator@example.com") {
		t.Fatalf("override recipient missing:\n%s", out)
	}
}

func TestNoSMTPLogsPayload(t *testing.T) {
	// No SMTP host: must not error; payload is logged.
	m := New("", 587, "", "", "kweeks@example.com", "", nil)
	if err := m.SendRedemptionEmail(context.Background(), "zainab@x.com", "KWEKS-ABC", "15000"); err != nil {
		t.Fatalf("expected nil err without SMTP, got %v", err)
	}
}

func hostOf(addr string) string {
	h, _, _ := net.SplitHostPort(addr)
	return h
}

func portOf(addr string) int {
	_, p, _ := net.SplitHostPort(addr)
	var n int
	_, _ = fmt.Sscanf(p, "%d", &n)
	return n
}
