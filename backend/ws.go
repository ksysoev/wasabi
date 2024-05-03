package backend

import (
	"bytes"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type WSBackend struct {
	URL    string
	Origin string
}

func (b *WSBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	c, _, err := websocket.Dial(conn.Context(), b.URL, nil)

	if err != nil {
		return err
	}

	defer c.CloseNow()

	err = c.Write(r.Context(), websocket.MessageText, r.Data())

	if err != nil {
		return err
	}

	msgType, reader, err := c.Reader(r.Context())
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(make([]byte, 0))
	_, err = buffer.ReadFrom(reader)

	if err != nil {
		return err
	}

	return conn.Send(msgType, buffer.Bytes())
}
