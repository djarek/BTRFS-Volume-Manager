package storageservers

import (
	"sync"

	"github.com/djarek/btrfs-volume-manager/common/dtos"
	"github.com/djarek/btrfs-volume-manager/common/request"
)

type serverMap map[dtos.StorageServerID]*request.Context

type serverTracker struct {
	serverMap serverMap
	mtx       sync.RWMutex
	nextID    dtos.StorageServerID
}

/*Tracker tracks storage servers currently connected to the system. */
type Tracker interface {
	GetServerContext(ID dtos.StorageServerID) (ctx *request.Context, ok bool)
	GetAllServers() []*request.Context
	RegisterServer(ctx *request.Context) dtos.StorageServerID
	RemoveServer(ID dtos.StorageServerID)
}

/*NewTracker constructs a new valid ServerTracker*/
func NewTracker() Tracker {
	return &serverTracker{serverMap: make(serverMap)}
}

func (s *serverTracker) GetServerContext(ID dtos.StorageServerID) (ctx *request.Context, ok bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	ctx, ok = s.serverMap[ID]
	return
}

func (s *serverTracker) GetAllServers() []*request.Context {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	var ctxList []*request.Context
	for _, ctx := range s.serverMap {
		ctxList = append(ctxList, ctx)
	}
	return ctxList
}

func (s *serverTracker) RegisterServer(ctx *request.Context) dtos.StorageServerID {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	ID := s.nextID
	s.nextID++
	s.serverMap[ID] = ctx
	return ID
}

func (s *serverTracker) RemoveServer(ID dtos.StorageServerID) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.serverMap, ID)
}
