package session

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/rgynn/dice/pkg/config"
)

var ErrMaxNumSessionsReached = errors.New("max number of sessions reached")
var ErrNotFound = errors.New("session not found")
var ErrNotEnoughPlayers = errors.New("not enough players to start session")
var ErrPlayerAlreadyRolled = errors.New("player already rolled dice for this session")

type Session struct {
	ID         string               `json:"id"`
	NumPlayers int                  `json:"num_players"`
	Highest    Roll                 `json:"-"`
	Timer      *time.Timer          `json:"-"`
	Done       chan struct{}        `json:"-"`
	Rolls      chan Roll            `json:"-"`
	Players    map[string]chan Roll `json:"-"`
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

type Roll struct {
	PlayerID string `json:"player_id"`
	Roll     int    `json:"roll"`
}

type Service struct {
	MaxNumSessions int
	MaxRollNumber  int
	Sessions       map[string]*Session
	CloseC         chan string
	sync.Mutex
}

func NewService(cfg *config.Data) (*Service, error) {
	return &Service{
		MaxNumSessions: cfg.MaxNumSessions,
		MaxRollNumber:  cfg.MaxRollNumber,
		Sessions:       map[string]*Session{},
		CloseC:         make(chan string, 1),
	}, nil
}

func (svc *Service) Run() {
	for sessionID := range svc.CloseC {
		svc.Lock()
		delete(svc.Sessions, sessionID)
		svc.Unlock()
	}
}

func (svc *Service) New(numPlayers, durationSeconds int) (*Session, error) {

	svc.Lock()
	defer svc.Unlock()

	if durationSeconds == 0 {
		durationSeconds = 10
	}

	if len(svc.Sessions) >= svc.MaxNumSessions {
		return nil, ErrMaxNumSessionsReached
	}

	if numPlayers < 2 {
		return nil, ErrNotEnoughPlayers
	}

	id := svc.newSessionID(20)

	sess := &Session{
		ID:         id,
		NumPlayers: numPlayers,
		Timer:      time.NewTimer(time.Duration(durationSeconds) * time.Second),
		Players:    map[string]chan Roll{},
		Rolls:      make(chan Roll, numPlayers),
		Done:       make(chan struct{}),
	}

	go sess.Open(svc.CloseC)

	svc.Sessions[id] = sess

	return sess, nil
}

func (svc *Service) AddRoll(sessionID, playerID string) (chan Roll, *Roll, error) {

	svc.Lock()
	defer svc.Unlock()

	sess, ok := svc.Sessions[sessionID]
	if !ok {
		return nil, nil, ErrNotFound
	}

	_, ok = sess.Players[playerID]
	if ok {
		return nil, nil, ErrPlayerAlreadyRolled
	}

	sess.Players[playerID] = make(chan Roll, 1)

	roll := Roll{
		PlayerID: playerID,
		Roll:     rand.Intn(svc.MaxRollNumber),
	}

	sess.Rolls <- roll

	if len(sess.Players) >= sess.NumPlayers {
		sess.Done <- struct{}{}
	}

	return sess.Players[playerID], &roll, nil
}

func (svc *Service) newSessionID(n int) string {
	new := randomString(n)
	for k := range svc.Sessions {
		if k == new {
			return svc.newSessionID(n)
		}
	}
	return new
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
