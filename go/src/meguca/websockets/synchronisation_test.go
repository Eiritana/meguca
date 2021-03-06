package websockets

import (
	"database/sql"
	"meguca/auth"
	"meguca/common"
	"meguca/db"
	"meguca/imager/assets"
	. "meguca/test"
	"meguca/websockets/feeds"
	"strconv"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOldFeedClosing(t *testing.T) {
	feeds.Clear()
	assertTableClear(t, "boards")
	writeSampleBoard(t)
	writeSampleThread(t)

	sv := newWSServer(t)
	defer sv.Close()
	cl, _ := sv.NewClient()
	registerClient(t, cl, 1, "a")

	err := cl.synchronise(marshalJSON(t, syncRequest{
		Thread: 0,
		Board:  "a",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if cl.feed != nil {
		t.Fatal("old feed not cleared")
	}
}

func TestSyncToBoard(t *testing.T) {
	feeds.Clear()
	setBoardConfigs(t, false)

	sv := newWSServer(t)
	defer sv.Close()
	cl, wcl := sv.NewClient()

	// Invalid board
	msg := syncRequest{
		Thread: 0,
		Board:  "c",
	}
	if err := cl.synchronise(marshalJSON(t, msg)); err != errInvalidBoard {
		UnexpectedError(t, err)
	}

	// Valid synchronization
	msg.Board = "a"
	if err := cl.synchronise(marshalJSON(t, msg)); err != nil {
		t.Fatal(err)
	}
	assertMessage(t, wcl, "30null")
}

func skipMessage(t *testing.T, con *websocket.Conn) {
	t.Helper()
	_, _, err := con.ReadMessage()
	if err != nil {
		t.Error(err)
	}
}

func TestRegisterSync(t *testing.T) {
	feeds.Clear()
	assertTableClear(t, "boards")
	writeSampleBoard(t)
	writeSampleThread(t)

	sv := newWSServer(t)
	defer sv.Close()
	cl, _ := sv.NewClient()

	syncs := [...]struct {
		id    uint64
		board string
	}{
		{1, "a"},
		{0, "a"},
	}

	// Both for new syncs and switching syncs
	for _, s := range syncs {
		err := cl.registerSync(syncRequest{
			Thread: s.id,
			Board:  s.board,
		})
		if err != nil {
			t.Fatal(err)
		}
		assertSyncID(t, cl, s.id, s.board)
	}
}

func assertSyncID(t *testing.T, cl *Client, id uint64, board string) {
	t.Helper()

	synced, _id, _board := feeds.GetSync(cl)
	if !synced {
		t.Error("client not synced")
	}
	if id != _id {
		LogUnexpected(t, id, _id)
	}
	if board != _board {
		LogUnexpected(t, board, _board)
	}
}

func TestInvalidThreadSync(t *testing.T) {
	feeds.Clear()
	assertTableClear(t, "boards")
	writeSampleBoard(t)

	sv := newWSServer(t)
	defer sv.Close()
	cl, _ := sv.NewClient()

	data := marshalJSON(t, syncRequest{
		Board:  "a",
		Thread: 1,
	})
	if err := cl.synchronise(data); err != errInvalidThread {
		UnexpectedError(t, err)
	}
}

func TestSyncToThread(t *testing.T) {
	feeds.Clear()
	assertTableClear(t, "boards")
	writeSampleBoard(t)
	writeSampleThread(t)

	sv := newWSServer(t)
	defer sv.Close()
	cl, wcl := sv.NewClient()
	sv.Add(1)
	go readListenErrors(t, cl, sv)

	sendMessage(t, wcl, common.MessageSynchronise, syncRequest{
		Board:  "a",
		Thread: 1,
	})

	skipMessage(t, wcl)
	skipMessage(t, wcl)
	assertMessage(t, wcl, "33[\"351\"]")
	assertSyncID(t, cl, 1, "a")

	cl.Close(nil)
	sv.Wait()
}

func sendMessage(
	t *testing.T,
	conn *websocket.Conn,
	typ common.MessageType,
	data interface{},
) {
	t.Helper()

	err := conn.WriteMessage(websocket.TextMessage, encodeMessage(t, typ, data))
	if err != nil {
		t.Fatal(err)
	}
}

func encodeMessage(
	t *testing.T,
	typ common.MessageType,
	data interface{},
) []byte {
	t.Helper()

	msg, err := common.EncodeMessage(typ, data)
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func TestReclaimPost(t *testing.T) {
	feeds.Clear()
	assertTableClear(t, "boards")
	writeSampleBoard(t)
	writeSampleThread(t)

	const pw = "123"
	hash, err := auth.BcryptHash(pw, 6)
	if err != nil {
		t.Fatal(err)
	}
	posts := [...]db.Post{
		{
			StandalonePost: common.StandalonePost{
				Post: common.Post{
					Editing: true,
					Image:   &assets.StdJPEG,
					ID:      2,
					Body:    "abc\ndef",
					Time:    3,
				},
				OP:    1,
				Board: "a",
			},
			Password: hash,
		},
		{
			StandalonePost: common.StandalonePost{
				Post: common.Post{
					Editing: false,
					ID:      3,
				},
				OP:    1,
				Board: "a",
			},
			Password: hash,
		},
	}
	err = db.InTransaction(func(tx *sql.Tx) error {
		for _, p := range posts {
			err := db.WritePost(tx, p, false, false)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := [...]struct {
		name     string
		id       uint64
		password string
		code     int
	}{
		{"no post", 99, "", 1},
		{"already closed", 3, "", 1},
		{"wrong password", 2, "aaaaaaaa", 1},
		{"valid", 2, pw, 0},
	}

	for i := range cases {
		c := cases[i]
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			sv := newWSServer(t)
			defer sv.Close()
			cl, wcl := sv.NewClient()
			registerClient(t, cl, 1, "a")
			req := reclaimRequest{
				ID:       c.id,
				Password: c.password,
			}
			if err := cl.reclaimPost(marshalJSON(t, req)); err != nil {
				t.Fatal(err)
			}

			assertMessage(t, wcl, `31`+strconv.Itoa(c.code))
		})
	}
}
