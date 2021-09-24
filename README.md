# Dice rolling service

Experiment with go channels. Made to resemble dice rolling mechanic in MMO's. Start a session with a given number of players. Wait for dice rolls from all players (or deadline exceeded), and return your roll and the winning roll.

## Requirements

### .env file

```
DEBUG=true
HOST=0.0.0.0
PORT=3000
MAX_NUM_SESSIONS=10
MAX_ROLL_NUM=100
```

## CLI Usage Example

### Start server

```
go run cmd/server/main.go
```

### Start session

```
DICE_SESSION_ID=$(go run cmd/client/main.go new)
```

### Roll dice

```
go run cmd/client/main.go roll --user $USER --session $DICE_SESSION_ID
```

## REST API

### Create session
```
curl -XPOST 'http://localhost:3000/sessions -d '{ "num_players": 2, "duration_seconds": 10 }'
```
### Roll dice
```
curl -XPOST 'http://localhost:3000/sessions/{sessionID}/{playerID}'
```