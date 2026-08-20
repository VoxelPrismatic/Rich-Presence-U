package discord

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const (
	opHandshake = 0
	opFrame     = 1
	opClose     = 2
	opPing      = 3
	opPong      = 4

	rpcVersion   = 1
	dialTimeout  = 500 * time.Millisecond
	writeTimeout = 5 * time.Second
	pingInterval = 5 * time.Second
	pingTimeout  = 10 * time.Second
	replyTimeout = 8 * time.Second
)

var (
	ErrNotConnected = errors.New("discord: not connected")
	ErrNoClient     = errors.New("discord: desktop client not found")
)

// User is the Discord account attached to the local client.
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

func (u User) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

func (u User) AvatarURL() string {
	if u.ID == "" || u.Avatar == "" {
		return ""
	}
	return "https://cdn.discordapp.com/avatars/" + u.ID + "/" + u.Avatar + ".png?size=128"
}

type frame struct {
	Op   uint32
	Body map[string]any
	Raw  []byte
}

// Client talks Discord IPC (SET_ACTIVITY) over a local pipe/socket.
type Client struct {
	mu       sync.Mutex
	conn     net.Conn
	clientID string
	user     User
	replies  map[string]chan frame
	ready    chan frame
	pong     chan struct{}
	closed   chan struct{}
	onClose  func(error)
}

func New() *Client {
	return &Client{}
}

func (c *Client) OnClose(fn func(error)) {
	c.mu.Lock()
	c.onClose = fn
	c.mu.Unlock()
}

func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

func (c *Client) User() User {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.user
}

func (c *Client) ClientID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clientID
}

// Connect handshakes with the first available Discord IPC pipe.
func (c *Client) Connect(ctx context.Context, clientID string) (User, error) {
	c.Close()
	var last error
	for _, path := range ipcPaths() {
		select {
		case <-ctx.Done():
			return User{}, ctx.Err()
		default:
		}
		conn, err := dialIPC(path, dialTimeout)
		if err != nil {
			last = err
			continue
		}
		user, err := c.handshake(ctx, conn, clientID)
		if err != nil {
			conn.Close()
			last = err
			continue
		}
		return user, nil
	}
	if last == nil {
		last = ErrNoClient
	}
	return User{}, fmt.Errorf("%w: %v", ErrNoClient, last)
}

func (c *Client) handshake(ctx context.Context, conn net.Conn, clientID string) (User, error) {
	closed := make(chan struct{})
	c.mu.Lock()
	c.conn = conn
	c.clientID = clientID
	c.user = User{}
	c.replies = map[string]chan frame{}
	c.ready = make(chan frame, 1)
	c.pong = make(chan struct{}, 1)
	c.closed = closed
	c.mu.Unlock()

	go c.readLoop()
	go c.pingLoop()

	if err := c.write(opHandshake, map[string]any{
		"v":         rpcVersion,
		"client_id": clientID,
	}); err != nil {
		c.Close()
		return User{}, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, replyTimeout)
	defer cancel()
	select {
	case <-waitCtx.Done():
		c.Close()
		return User{}, waitCtx.Err()
	case <-closed:
		c.Close()
		return User{}, ErrNotConnected
	case f := <-c.ready:
		user, err := parseUser(f.Body)
		if err != nil {
			c.Close()
			return User{}, err
		}
		c.mu.Lock()
		c.user = user
		c.mu.Unlock()
		return user, nil
	}
}

func parseUser(body map[string]any) (User, error) {
	data, _ := body["data"].(map[string]any)
	if data == nil {
		data = body
	}
	raw, _ := data["user"].(map[string]any)
	if raw == nil {
		raw, _ = body["user"].(map[string]any)
	}
	if raw == nil {
		return User{}, errors.New("discord: handshake missing user")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return User{}, err
	}
	var u User
	if err := json.Unmarshal(b, &u); err != nil {
		return User{}, err
	}
	if u.ID == "" {
		return User{}, errors.New("discord: handshake user has no id")
	}
	return u, nil
}

// SetActivity pushes or clears (nil) the current rich presence.
func (c *Client) SetActivity(ctx context.Context, a *Activity) error {
	if !c.Connected() {
		return ErrNotConnected
	}
	args := map[string]any{"pid": os.Getpid()}
	if a != nil {
		args["activity"] = a.payload()
	} else {
		args["activity"] = nil
	}
	nonce := newNonce()
	wait := make(chan frame, 1)
	c.mu.Lock()
	if c.replies == nil {
		c.mu.Unlock()
		return ErrNotConnected
	}
	c.replies[nonce] = wait
	c.mu.Unlock()
	defer c.drop(nonce)

	if err := c.write(opFrame, map[string]any{
		"cmd":   "SET_ACTIVITY",
		"args":  args,
		"nonce": nonce,
	}); err != nil {
		return err
	}
	waitCtx, cancel := context.WithTimeout(ctx, replyTimeout)
	defer cancel()
	select {
	case <-waitCtx.Done():
		return waitCtx.Err()
	case f := <-wait:
		if evt, _ := f.Body["evt"].(string); evt == "ERROR" {
			return fmt.Errorf("discord: %v", f.Body["data"])
		}
		return nil
	}
}

func (c *Client) Clear(ctx context.Context) error {
	return c.SetActivity(ctx, nil)
}

func (c *Client) drop(nonce string) {
	c.mu.Lock()
	delete(c.replies, nonce)
	c.mu.Unlock()
}

func (c *Client) Close() {
	c.mu.Lock()
	conn := c.conn
	closed := c.closed
	c.conn = nil
	c.clientID = ""
	c.user = User{}
	c.replies = nil
	c.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
	if closed != nil {
		select {
		case <-closed:
		default:
			close(closed)
		}
	}
}

func (c *Client) write(op uint32, body map[string]any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var hdr [8]byte
	binary.LittleEndian.PutUint32(hdr[0:4], op)
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(raw)))

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return ErrNotConnected
	}
	_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if _, err := conn.Write(hdr[:]); err != nil {
		c.fail(err)
		return err
	}
	if _, err := conn.Write(raw); err != nil {
		c.fail(err)
		return err
	}
	return nil
}

func (c *Client) readLoop() {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return
	}
	for {
		f, err := readFrame(conn)
		if err != nil {
			c.fail(err)
			return
		}
		c.dispatch(f)
	}
}

func (c *Client) pingLoop() {
	c.mu.Lock()
	closed := c.closed
	pong := c.pong
	c.mu.Unlock()
	if closed == nil {
		return
	}
	t := time.NewTicker(pingInterval)
	defer t.Stop()
	for {
		select {
		case <-closed:
			return
		case <-t.C:
			if err := c.write(opPing, map[string]any{"nonce": newNonce()}); err != nil {
				return
			}
			select {
			case <-closed:
				return
			case <-pong:
			case <-time.After(pingTimeout):
				c.fail(errors.New("discord: ping timeout"))
				return
			}
		}
	}
}

func (c *Client) dispatch(f frame) {
	switch f.Op {
	case opClose:
		c.fail(errors.New("discord: ipc closed"))
		return
	case opPing:
		_ = c.write(opPong, f.Body)
		return
	case opPong:
		c.mu.Lock()
		ch := c.pong
		c.mu.Unlock()
		if ch != nil {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		return
	}

	nonce, _ := f.Body["nonce"].(string)
	evt, _ := f.Body["evt"].(string)
	cmd, _ := f.Body["cmd"].(string)
	ready := evt == "READY" || (cmd == "DISPATCH" && evt == "READY") || f.Op == opHandshake
	if !ready {
		if data, ok := f.Body["data"].(map[string]any); ok {
			if _, ok := data["user"]; ok {
				ready = true
			}
		}
	}

	c.mu.Lock()
	var dest chan frame
	if nonce != "" && c.replies != nil {
		dest = c.replies[nonce]
	}
	readyCh := c.ready
	c.mu.Unlock()

	if dest != nil {
		select {
		case dest <- f:
		default:
		}
		return
	}
	if ready && readyCh != nil {
		select {
		case readyCh <- f:
		default:
		}
	}
}

func (c *Client) fail(err error) {
	c.mu.Lock()
	fn := c.onClose
	conn := c.conn
	c.conn = nil
	closed := c.closed
	c.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
	if closed != nil {
		select {
		case <-closed:
		default:
			close(closed)
		}
	}
	if fn != nil && err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
		fn(err)
	}
}

func readFrame(r io.Reader) (frame, error) {
	var hdr [8]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return frame{}, err
	}
	op := binary.LittleEndian.Uint32(hdr[0:4])
	n := binary.LittleEndian.Uint32(hdr[4:8])
	if n > 16*1024*1024 {
		return frame{}, fmt.Errorf("discord: frame too large (%d)", n)
	}
	body := make([]byte, n)
	if n > 0 {
		if _, err := io.ReadFull(r, body); err != nil {
			return frame{}, err
		}
	}
	out := frame{Op: op, Raw: body, Body: map[string]any{}}
	if n > 0 {
		if err := json.Unmarshal(body, &out.Body); err != nil {
			return frame{}, err
		}
	}
	return out, nil
}

func newNonce() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
