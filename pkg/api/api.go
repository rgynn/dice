package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rgynn/dice/pkg/session"
	"github.com/rgynn/dice/pkg/session/local"
)

type Service struct {
	sessions session.Keeper
}

func NewService(maxNumSessions, maxRollNumber int) (*Service, error) {
	sessions, err := local.NewKeeper(maxNumSessions, maxRollNumber)
	if err != nil {
		return nil, err
	}
	go sessions.Run()
	return &Service{
		sessions: sessions,
	}, nil
}

func (svc *Service) NewSessionHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		NumPlayers      int `json:"num_players"`
		DurationSeconds int `json:"duration_seconds"`
	}
	reqbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()
	var req request
	if err := json.Unmarshal(reqbody, &req); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}
	sess, err := svc.sessions.NewSession(r.Context(), req.NumPlayers, req.DurationSeconds)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	body, err := json.Marshal(sess)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	NewResponse(w, r, http.StatusOK, body)
}

func (svc *Service) NewRollHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionID"]
	if sessionID == "" {
		NewErrorResponse(w, r, http.StatusBadRequest, errors.New("no sessionID provided"))
		return
	}
	playerID := mux.Vars(r)["playerID"]
	if playerID == "" {
		NewErrorResponse(w, r, http.StatusBadRequest, errors.New("no playerID provided"))
		return
	}
	resultC, roll, err := svc.sessions.AddSessionRoll(r.Context(), sessionID, playerID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	result := <-resultC
	type response struct {
		Your   *session.Roll `json:"your"`
		Winner session.Roll  `json:"winner"`
	}
	body, err := json.Marshal(&response{Your: roll, Winner: result})
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	NewResponse(w, r, http.StatusOK, body)
}

func NewResponse(w http.ResponseWriter, r *http.Request, status int, body []byte) {
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		return
	}
}

func NewErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	type ErrorResponse struct {
		Path    string `json:"path"`
		Method  string `json:"method"`
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	body, merr := json.Marshal(&ErrorResponse{
		Path:    r.URL.Path,
		Method:  r.Method,
		Code:    status,
		Message: err.Error(),
	})
	if merr != nil {
		return
	}
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		return
	}
}
