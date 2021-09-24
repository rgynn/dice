package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Data struct {
	LogLevel       logrus.Level
	Port           int
	Host           string
	Addr           string
	MaxNumSessions int
	MaxRollNumber  int
}

func NewFromEnv(filenames ...string) (*Data, error) {
	if err := godotenv.Load(filenames...); err != nil {
		log.Printf("WARNING: %s", err.Error())
	}
	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env variable DEBUG: %w", err)
	}
	var logLevel logrus.Level
	if debug {
		logLevel = logrus.DebugLevel
	} else {
		logLevel = logrus.WarnLevel
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env variable PORT: %w", err)
	}
	host := os.Getenv("HOST")
	if host == "" {
		return nil, errors.New("failed to read env variable HOST")
	}
	maxNumSessions, err := strconv.Atoi(os.Getenv("MAX_NUM_SESSIONS"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env variable MAX_NUM_SESSIONS: %w", err)
	}
	maxRollNumber, err := strconv.Atoi(os.Getenv("MAX_ROLL_NUM"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env variable MAX_ROLL_NUM: %w", err)
	}
	return &Data{
		LogLevel:       logLevel,
		Port:           port,
		Host:           host,
		Addr:           fmt.Sprintf("%s:%d", host, port),
		MaxNumSessions: maxNumSessions,
		MaxRollNumber:  maxRollNumber,
	}, nil
}
