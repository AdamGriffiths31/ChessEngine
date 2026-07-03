// Package benchmark provides interactive chess engine benchmarking functionality.
package benchmark

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// StockfishClient drives a Stockfish (or any UCI engine) subprocess over
// stdin/stdout. A background goroutine pumps stdout lines into a channel so
// that WaitFor can respect a context deadline instead of blocking forever
// on a hung or crashed process.
type StockfishClient struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	lines   chan string
	done    chan struct{}
	traffic trafficLog
}

// trafficLog is a fixed-size ring buffer of timestamped UCI traffic kept for
// failure diagnostics: commands sent ("->"), engine stdout ("<-") and engine
// stderr ("!!"). Safe for concurrent use by the send path and pump goroutines.
type trafficLog struct {
	mu      sync.Mutex
	entries [256]string
	next    int
	count   int
}

func (t *trafficLog) add(direction, text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[t.next] = time.Now().Format("15:04:05.000") + " " + direction + " " + text
	t.next = (t.next + 1) % len(t.entries)
	if t.count < len(t.entries) {
		t.count++
	}
}

// tail returns up to maxLines of the most recent traffic, oldest first.
func (t *trafficLog) tail(maxLines int) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	n := t.count
	if maxLines > 0 && n > maxLines {
		n = maxLines
	}
	out := make([]string, 0, n)
	start := t.next - n
	if start < 0 {
		start += len(t.entries)
	}
	for i := 0; i < n; i++ {
		out = append(out, t.entries[(start+i)%len(t.entries)])
	}
	return out
}

// NewStockfishClient spawns the engine binary at binaryPath and starts
// pumping its stdout.
func NewStockfishClient(binaryPath string) (*StockfishClient, error) {
	cmd := exec.Command(binaryPath) // #nosec G204 - binary path resolved from local tools/engines.json

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine at %s: %w", binaryPath, err)
	}

	client := &StockfishClient{
		cmd:   cmd,
		stdin: stdin,
		lines: make(chan string, 256),
		done:  make(chan struct{}),
	}

	go client.pumpLines(stdout)
	go client.pumpStderr(stderr)

	return client, nil
}

// pumpLines reads stdout line by line and forwards each line to c.lines
// until stdout closes or Close() is called.
func (c *StockfishClient) pumpLines(stdout io.ReadCloser) {
	defer close(c.lines)
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		c.traffic.add("<-", line)
		select {
		case c.lines <- line:
		case <-c.done:
			return
		}
	}
	if err := scanner.Err(); err != nil {
		c.traffic.add("!!", fmt.Sprintf("stdout scanner error: %v", err))
	}
}

// pumpStderr records engine stderr into the traffic log. Engines rarely write
// to stderr, but crash messages land there and are gold when the engine goes
// silent instead of answering.
func (c *StockfishClient) pumpStderr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		c.traffic.add("!!", scanner.Text())
	}
}

// Send writes a single UCI command, terminated with a newline.
func (c *StockfishClient) Send(command string) error {
	c.traffic.add("->", command)
	if _, err := fmt.Fprintf(c.stdin, "%s\n", command); err != nil {
		return fmt.Errorf("failed to send command %q: %w", command, err)
	}
	return nil
}

// TrafficDump returns the most recent UCI traffic formatted one entry per
// line, oldest first: "->" commands sent, "<-" engine stdout, "!!" engine
// stderr and pump errors. maxLines <= 0 returns everything retained.
func (c *StockfishClient) TrafficDump(maxLines int) string {
	return strings.Join(c.traffic.tail(maxLines), "\n")
}

// WaitFor blocks until a line starting with prefix arrives, ctx is done,
// or the engine's stdout closes. It returns the matching line.
func (c *StockfishClient) WaitFor(ctx context.Context, prefix string) (string, error) {
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return "", fmt.Errorf("engine process closed stdout before sending %q", prefix)
			}
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("timed out waiting for %q: %w", prefix, ctx.Err())
		}
	}
}

// WaitForAll behaves like WaitFor but returns every line seen, including
// the matching terminal line. Used to scan UCI option declarations during
// the "uci" handshake.
func (c *StockfishClient) WaitForAll(ctx context.Context, prefix string) ([]string, error) {
	var collected []string
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return collected, fmt.Errorf("engine process closed stdout before sending %q", prefix)
			}
			collected = append(collected, line)
			if strings.HasPrefix(line, prefix) {
				return collected, nil
			}
		case <-ctx.Done():
			return collected, fmt.Errorf("timed out waiting for %q: %w", prefix, ctx.Err())
		}
	}
}

// Close asks the engine to quit and releases the process. If the process
// doesn't exit within 5 seconds of being asked to quit, it is killed
// forcibly so a wedged engine can't hang the caller indefinitely.
func (c *StockfishClient) Close() error {
	close(c.done)
	_ = c.Send("quit")
	_ = c.stdin.Close()

	waitErr := make(chan error, 1)
	go func() { waitErr <- c.cmd.Wait() }()

	select {
	case err := <-waitErr:
		return err
	case <-time.After(5 * time.Second):
		_ = c.cmd.Process.Kill()
		<-waitErr // Wait() always returns once the process is gone, killed or not.
		return fmt.Errorf("engine did not exit after quit; killed process")
	}
}
