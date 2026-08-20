package discord

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func writeFrame(w io.Writer, op uint32, body any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var hdr [8]byte
	binary.LittleEndian.PutUint32(hdr[0:4], op)
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(raw)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err = w.Write(raw)
	return err
}

func TestConnectSetActivity(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "discord-ipc-0")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("TMPDIR", dir)

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		f, err := readFrame(conn)
		if err != nil {
			t.Error(err)
			return
		}
		if f.Op != opHandshake {
			t.Errorf("op %d", f.Op)
		}
		if f.Body["client_id"] != "app" {
			t.Errorf("client_id %v", f.Body["client_id"])
		}
		_ = writeFrame(conn, opFrame, map[string]any{
			"cmd": "DISPATCH",
			"evt": "READY",
			"data": map[string]any{
				"user": map[string]any{
					"id":          "1",
					"username":    "ninstar",
					"global_name": "NinStar",
					"avatar":      "abc",
				},
			},
		})
		f, err = readFrame(conn)
		if err != nil {
			t.Error(err)
			return
		}
		if f.Body["cmd"] != "SET_ACTIVITY" {
			t.Errorf("cmd %v", f.Body["cmd"])
		}
		nonce, _ := f.Body["nonce"].(string)
		args, _ := f.Body["args"].(map[string]any)
		act, _ := args["activity"].(map[string]any)
		if act["details"] != "Hello" {
			t.Errorf("activity %+v", act)
		}
		_ = writeFrame(conn, opFrame, map[string]any{
			"cmd":   "SET_ACTIVITY",
			"nonce": nonce,
			"data":  map[string]any{},
		})
	}()

	cli := New()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	user, err := cli.Connect(ctx, "app")
	if err != nil {
		t.Fatal(err)
	}
	if user.DisplayName() != "NinStar" || user.AvatarURL() == "" {
		t.Fatalf("user %+v", user)
	}
	act := Build(Presence{Title: "Hello", CoverKey: "default"})
	if err := cli.SetActivity(ctx, &act); err != nil {
		t.Fatal(err)
	}
	cli.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("server hung")
	}
}

func TestConnectMissing(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	t.Setenv("TMPDIR", t.TempDir())
	cli := New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := cli.Connect(ctx, "app")
	if err == nil {
		t.Fatal("expected error")
	}
}
