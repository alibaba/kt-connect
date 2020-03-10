package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

// StreamHandler ...
type StreamHandler struct {
	WsConn      *Connection
	ResizeEvent chan remotecommand.TerminalSize
}

type xtermMessage struct {
	MsgType string `json:"type"` // resize || input
	Input   string `json:"input"`
	Rows    uint16 `json:"rows"`
	Cols    uint16 `json:"cols"`
}

// Next ...
func (handler *StreamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.ResizeEvent
	size = &ret
	return
}

// Read ...
func (handler *StreamHandler) Read(p []byte) (size int, err error) {
	var (
		msg      *Message
		xtermMsg xtermMessage
	)

	if msg, err = handler.WsConn.WsRead(); err != nil {
		return
	}

	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		return
	}

	if xtermMsg.MsgType == "resize" {
		handler.ResizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	} else if xtermMsg.MsgType == "input" {
		size = len(xtermMsg.Input)
		copy(p, xtermMsg.Input)
	}
	return
}

// Write ...
func (handler *StreamHandler) Write(p []byte) (size int, err error) {
	var (
		copyData []byte
	)

	copyData = make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.WsConn.WsWrite(websocket.TextMessage, copyData)
	return
}
