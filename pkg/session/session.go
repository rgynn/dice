package session

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Keeper interface {
	NewSession(ctx context.Context, maxNumPlayers, maxDurationSeconds int) (*Session, error)
	AddSessionRoll(ctx context.Context, sessionID, playerID string) (chan Roll, *Roll, error)
	Run()
}

var ErrMaxNumSessionsReached = errors.New("max number of sessions reached")
var ErrMaxNumPlayersReached = errors.New("max number of players for this session reached")
var ErrNotFound = errors.New("session not found")
var ErrNotEnoughPlayers = errors.New("not enough players to start session")
var ErrPlayerAlreadyRolled = errors.New("player already rolled dice for this session")

type Roll struct {
	PlayerID string `json:"player_id"`
	Roll     int    `json:"roll"`
}

type Session struct {
	ID            string               `json:"id"`
	MaxNumPlayers int                  `json:"num_players"`
	Highest       Roll                 `json:"-"`
	Timer         *time.Timer          `json:"-"`
	Done          chan struct{}        `json:"-"`
	Rolls         chan Roll            `json:"-"`
	Players       map[string]chan Roll `json:"-"`
	sync.Mutex
}

func (sess *Session) Open(closeC chan string) {
	defer func() {
		sess.Close(closeC)
	}()
	for {
		select {
		case <-sess.Done:
			return
		case <-sess.Timer.C:
			return
		case roll := <-sess.Rolls:
			if roll.Roll > sess.Highest.Roll {
				sess.Highest = roll
			}
		}
	}
}

func (sess *Session) Close(closeC chan string) {
	for _, resultC := range sess.Players {
		resultC <- sess.Highest
	}
	closeC <- sess.ID
}

func (sess *Session) AddRoll(ctx context.Context, sessionID, playerID string, max int) (chan Roll, *Roll, error) {
	sess.Lock()
	defer sess.Unlock()
	if len(sess.Players) >= sess.MaxNumPlayers {
		return nil, nil, ErrMaxNumPlayersReached
	}
	_, ok := sess.Players[playerID]
	if ok {
		return nil, nil, ErrPlayerAlreadyRolled
	}
	sess.Players[playerID] = make(chan Roll)
	roll := Roll{
		PlayerID: playerID,
		Roll:     rand.Intn(max),
	}
	sess.Rolls <- roll
	if len(sess.Players) >= sess.MaxNumPlayers {
		sess.Done <- struct{}{}
	}
	return sess.Players[playerID], &roll, nil
}
