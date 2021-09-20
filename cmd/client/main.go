package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rgynn/dice/pkg/session"
	"github.com/spf13/cobra"
)

type Client struct {
	Username        *string
	URL             *string
	SessionID       *string
	NumPlayers      *int
	DurationSeconds *int
	http.Client
}

var client Client

var rootcmd = &cobra.Command{
	Use: "dice",
}

var newcmd = &cobra.Command{
	Use: "new",
	Run: newSession,
}

var rollcmd = &cobra.Command{
	Use: "roll",
	Run: roll,
}

func init() {

	client.URL = rootcmd.PersistentFlags().String("url", "http://localhost:3000", "url to dice rolling service")
	client.NumPlayers = newcmd.Flags().Int("num", 2, "number of players per session")
	client.DurationSeconds = newcmd.Flags().Int("duration", 10, "session duration in seconds")
	client.Username = rollcmd.Flags().String("user", "", "username, must be unique per session")
	client.SessionID = rollcmd.Flags().String("session", "", "session id to roll for")

	if err := rollcmd.MarkFlagRequired("user"); err != nil {
		log.Fatal(err)
	}

	if err := rollcmd.MarkFlagRequired("session"); err != nil {
		log.Fatal(err)
	}

	rootcmd.AddCommand(newcmd)
	rootcmd.AddCommand(rollcmd)
}

func main() {
	if err := rootcmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func newSession(cmd *cobra.Command, args []string) {

	type request struct {
		NumPlayers      *int `json:"num_players"`
		DurationSeconds *int `json:"duration_seconds"`
	}

	reqbody, err := json.Marshal(&request{NumPlayers: client.NumPlayers, DurationSeconds: client.DurationSeconds})
	if err != nil {
		log.Fatalf("failed to marshal new session request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sessions", *client.URL), bytes.NewReader(reqbody))
	if err != nil {
		log.Fatalf("Failed to create new session request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to create new session: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body when creating new session: %v", err)
	}
	defer resp.Body.Close()

	var sess session.Session
	if err := json.Unmarshal(body, &sess); err != nil {
		log.Fatalf("Failed to unmarshal response body when trying to create new session: %v", err)
	}

	fmt.Println(sess.ID)
}

func roll(cmd *cobra.Command, args []string) {

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sessions/%s/%s", *client.URL, *client.SessionID, *client.Username), nil)
	if err != nil {
		log.Fatalf("Failed to create new session request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to create new session: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body when creating new session: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		log.Fatal(string(body))
	}

	type respo struct {
		Your   session.Roll `json:"your"`
		Winner session.Roll `json:"winner"`
	}

	var response respo
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Failed to unmarshal response body from roll: %v", err)
	}

	if response.Winner.PlayerID == *client.Username {
		log.Printf("You won with: %d", response.Your.Roll)
	} else {
		log.Printf("%s won with: %d, you rolled: %d", response.Winner.PlayerID, response.Winner.Roll, response.Your.Roll)
	}
}
