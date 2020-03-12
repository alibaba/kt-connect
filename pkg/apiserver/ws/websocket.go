package ws

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message ...
type Message struct {
	MessageType int
	Data        []byte
}

// Connection ...
type Connection struct {
	wsSocket *websocket.Conn
	inChan   chan *Message
	outChan  chan *Message

	mutex     sync.Mutex
	isClosed  bool
	closeChan chan byte
}

// Constructor ...
func Constructor(resp http.ResponseWriter, req *http.Request) (wsConn *Connection, err error) {
	wsSocket, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}
	wsConn = &Connection{
		wsSocket:  wsSocket,
		inChan:    make(chan *Message, 1000),
		outChan:   make(chan *Message, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	go wsConn.wsReadLoop()
	go wsConn.wsWriteLoop()

	return
}

func (wsConn *Connection) wsReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *Message
		err     error
	)
	for {
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			goto ERROR
		}
		msg = &Message{
			msgType,
			data,
		}
		select {
		case wsConn.inChan <- msg:
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

func (wsConn *Connection) wsWriteLoop() {
	var (
		msg *Message
		err error
	)
	for {
		select {
		case msg = <-wsConn.outChan:
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

// WsClose ...
func (wsConn *Connection) WsClose() {
	wsConn.wsSocket.Close()

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}

// WsWrite ...
func (wsConn *Connection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &Message{messageType, data}:
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

// WsRead ...
func (wsConn *Connection) WsRead() (msg *Message, err error) {
	select {
	case msg = <-wsConn.inChan:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}
