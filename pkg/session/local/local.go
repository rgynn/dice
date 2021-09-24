package local

import (
	"context"
	"sync"
	"time"

	"github.com/rgynn/dice/pkg/helper"
	"github.com/rgynn/dice/pkg/session"
)

type Keeper struct {
	MaxNumSessions int
	MaxRollNumber  int
	Sessions       map[string]*session.Session
	NewC           chan *session.Session
	CloseC         chan string
	sync.Mutex
}

func NewKeeper(maxNumSessions, maxRandom int) (session.Keeper, error) {
	return &Keeper{
		MaxNumSessions: maxNumSessions,
		MaxRollNumber:  maxRandom,
		Sessions:       map[string]*session.Session{},
		NewC:           make(chan *session.Session, 1),
		CloseC:         make(chan string, 1),
	}, nil
}

func (svc *Keeper) NewSession(ctx context.Context, maxNumPlayers, maxDurationSeconds int) (*session.Session, error) {
	if maxDurationSeconds == 0 {
		maxDurationSeconds = 10
	}
	if len(svc.Sessions) >= svc.MaxNumSessions {
		return nil, session.ErrMaxNumSessionsReached
	}
	if maxNumPlayers < 2 {
		return nil, session.ErrNotEnoughPlayers
	}
	svc.Lock()
	id := svc.newSessionID(20)
	svc.Unlock()
	sess := &session.Session{
		ID:            id,
		MaxNumPlayers: maxNumPlayers,
		Timer:         time.NewTimer(time.Duration(maxDurationSeconds) * time.Second),
		Players:       map[string]chan session.Roll{},
		Rolls:         make(chan session.Roll, maxNumPlayers),
		Done:          make(chan struct{}),
	}
	go sess.Open(svc.CloseC)
	svc.NewC <- sess
	return sess, nil
}

func (svc *Keeper) AddSessionRoll(ctx context.Context, sessionID, playerID string) (chan session.Roll, *session.Roll, error) {
	svc.Lock()
	sess, ok := svc.Sessions[sessionID]
	svc.Unlock()
	if !ok {
		return nil, nil, session.ErrNotFound
	}
	return sess.AddRoll(ctx, sessionID, playerID, svc.MaxRollNumber)
}

func (svc *Keeper) Run() {
	for {
		select {
		case sess := <-svc.NewC:
			svc.storeSession(sess)
		case sessionID := <-svc.CloseC:
			svc.removeSession(sessionID)
		}
	}
}

func (svc *Keeper) newSessionID(n int) string {
	new := helper.RandomString(n)
	for k := range svc.Sessions {
		if k == new {
			return svc.newSessionID(n)
		}
	}
	return new
}

func (svc *Keeper) storeSession(sess *session.Session) {
	svc.Lock()
	svc.Sessions[sess.ID] = sess
	svc.Unlock()
}

func (svc *Keeper) removeSession(sessionID string) {
	svc.Lock()
	delete(svc.Sessions, sessionID)
	svc.Unlock()
}
