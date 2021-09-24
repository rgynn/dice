package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rgynn/dice/pkg/session"
)

type mockKeeper struct {
	NewSesssionFunc    func(ctx context.Context, maxNumPlayers, maxDurationSeconds int) (*session.Session, error)
	AddSessionRollFunc func(ctx context.Context, sessionID, playerID string) (chan session.Roll, *session.Roll, error)
	RunFunc            func()
}

func (mock *mockKeeper) NewSession(ctx context.Context, maxNumPlayers, maxDurationSeconds int) (*session.Session, error) {
	return mock.NewSesssionFunc(ctx, maxNumPlayers, maxDurationSeconds)
}
func (mock *mockKeeper) AddSessionRoll(ctx context.Context, sessionID, playerID string) (chan session.Roll, *session.Roll, error) {
	return mock.AddSessionRollFunc(ctx, sessionID, playerID)
}
func (mock *mockKeeper) Run() {}

func TestService_NewSessionHandler(t *testing.T) {
	type testcase struct {
		Name           string
		Input          []byte
		ExpectedStatus int
		ExpectedBody   []byte
	}
	testcases := []testcase{
		{
			Name:           "Happy path",
			Input:          []byte(`{"num_players": 2,"duration_seconds": 10}`),
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   []byte(`{"id":"fakeid","num_players":2}`),
		},
		{
			Name:           "Invalid body",
			Input:          nil,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   []byte(`{"path":"/","method":"POST","code":400,"msg":"unexpected end of JSON input"}`),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.Input))
			w := httptest.NewRecorder()
			svc := &Service{
				sessions: &mockKeeper{
					NewSesssionFunc: func(ctx context.Context, maxNumPlayers, maxDurationSeconds int) (*session.Session, error) {
						return &session.Session{
							ID:            "fakeid",
							MaxNumPlayers: maxNumPlayers,
						}, nil
					},
				},
			}
			svc.NewSessionHandler(w, r)
			if want, got := tc.ExpectedStatus, w.Code; want != got {
				t.Errorf("expected http status code: %v, got: %v", want, got)
			}
			if !bytes.Equal(tc.ExpectedBody, w.Body.Bytes()) {
				t.Errorf("expected http response body: %s, got: %s", tc.ExpectedBody, w.Body.Bytes())
			}
		})
	}
}

func TestService_NewRollHandler(t *testing.T) {
	type testcase struct {
		Name           string
		InputSessionID string
		InputPlayerID  string
		ExpectedStatus int
		ExpectedBody   []byte
	}
	testcases := []testcase{
		{
			Name:           "Winning round",
			InputSessionID: "fakesession",
			InputPlayerID:  "winninguser",
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   []byte(`{"your":{"player_id":"winninguser","roll":100},"winner":{"player_id":"winninguser","roll":100}}`),
		},
		{
			Name:           "Losing round",
			InputSessionID: "fakesession",
			InputPlayerID:  "losinguser",
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   []byte(`{"your":{"player_id":"losinguser","roll":50},"winner":{"player_id":"otheruser","roll":100}}`),
		},
		{
			Name:           "No sessionID",
			InputSessionID: "",
			InputPlayerID:  "user",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   []byte(`{"path":"/","method":"POST","code":400,"msg":"no sessionID provided"}`),
		},
		{
			Name:           "No playerID",
			InputSessionID: "fakesession",
			InputPlayerID:  "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   []byte(`{"path":"/","method":"POST","code":400,"msg":"no playerID provided"}`),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()
			svc := &Service{
				sessions: &mockKeeper{
					AddSessionRollFunc: func(ctx context.Context, sessionID, playerID string) (chan session.Roll, *session.Roll, error) {
						rollC := make(chan session.Roll, 1)
						var roll, winningRoll session.Roll
						switch tc.InputPlayerID {
						case "winninguser":
							roll = session.Roll{PlayerID: "winninguser", Roll: 100}
							winningRoll = session.Roll{PlayerID: "winninguser", Roll: 100}
						case "losinguser":
							roll = session.Roll{PlayerID: "losinguser", Roll: 50}
							winningRoll = session.Roll{PlayerID: "otheruser", Roll: 100}
						}
						defer func() {
							rollC <- winningRoll
						}()
						return rollC, &roll, nil
					},
				},
			}
			r = mux.SetURLVars(r, map[string]string{
				"sessionID": tc.InputSessionID,
				"playerID":  tc.InputPlayerID,
			})
			svc.NewRollHandler(w, r)
			if want, got := tc.ExpectedStatus, w.Code; want != got {
				t.Errorf("expected http status code: %v, got: %v", want, got)
			}
			if !bytes.Equal(tc.ExpectedBody, w.Body.Bytes()) {
				t.Errorf("expected http response body: %s, got: %s", tc.ExpectedBody, w.Body.Bytes())
			}
		})
	}
}
