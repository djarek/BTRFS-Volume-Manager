package wsserver

import "sync"

//ConnectionManager is a thread safe websocket connection & session manager
type ConnectionManager struct {
	masterConnection Connection
	mtx              sync.Mutex
}
