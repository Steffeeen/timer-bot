package main

import (
	"database/sql"
	"errors"
	"math/rand/v2"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() error {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "/app/data/timerbot.db"
	}
	db, err = sql.Open("sqlite3", dbURL)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS timers (
			internalId INTEGER PRIMARY KEY AUTOINCREMENT,
			id TEXT UNIQUE,
			message TEXT,
			user TEXT,
			channel TEXT,
			creation DATETIME,
			due DATETIME,
			snoozedDue DATETIME,
			snoozeCount INTEGER DEFAULT 0,
			shown BOOLEAN DEFAULT false
		)
	`)

	return err
}

func createTimer(id string, message string, userId string, channelId string, due time.Time) (*Timer, error) {
	created := time.Now()
	_, err := db.Exec("INSERT INTO timers (id, message, user, channel, creation, due, snoozedDue) VALUES (?, ?, ?, ?, ?, ?, ?)", id, message, userId, channelId, created, due, due)
	if err != nil {
		return nil, err
	}

	return &Timer{
		ID:          id,
		Message:     message,
		User:        userId,
		Channel:     channelId,
		Created:     created,
		Due:         due,
		SnoozedDue:  due,
		SnoozeCount: 0,
		Shown:       false,
	}, nil
}

func getTimerByID(id string) (*Timer, error) {
	row := db.QueryRow("SELECT internalId, id, message, user, channel, creation, due, snoozedDue, snoozeCount, shown FROM timers WHERE id = ?", id)
	timer := &Timer{}

	err := row.Scan(&timer.InternalID, &timer.ID, &timer.Message, &timer.User, &timer.Channel, &timer.Created, &timer.Due, &timer.SnoozedDue, &timer.SnoozeCount, &timer.Shown)
	if err != nil {
		return nil, err
	}

	return timer, nil
}

func getAllTimersForUser(userID string, onlyActive bool) ([]*Timer, error) {
	query := "SELECT internalId, id, message, user, channel, creation, due, snoozedDue, snoozeCount, shown FROM timers WHERE user = ?"
	if onlyActive {
		query += " AND shown = false"
	}
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	var timers []*Timer
	for rows.Next() {
		timer := &Timer{}
		err := rows.Scan(&timer.InternalID, &timer.ID, &timer.Message, &timer.User, &timer.Channel, &timer.Created, &timer.Due, &timer.SnoozedDue, &timer.SnoozeCount, &timer.Shown)
		if err != nil {
			return nil, err
		}
		timers = append(timers, timer)
	}
	return timers, err
}

func updateTimer(timer *Timer) error {
	_, err := db.Exec("UPDATE timers SET message = ?, due = ?, snoozedDue = ? WHERE id = ?", timer.Message, timer.Due, timer.SnoozedDue, timer.ID)
	return err
}

func deleteTimer(id string) error {
	_, err := db.Exec("DELETE FROM timers WHERE id = ?", id)
	return err
}

func snoozeTimer(id string, newDueDate time.Time) error {
	_, err := db.Exec("UPDATE timers SET snoozedDue = ?, snoozeCount = snoozeCount + 1, shown = false WHERE id = ?", newDueDate, id)
	return err
}

func getDueTimers() ([]*Timer, error) {
	rows, err := db.Query("SELECT internalId, id, message, user, channel, creation, due, snoozedDue, snoozeCount, shown FROM timers WHERE snoozedDue <= ? AND shown = false", time.Now())
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	var timers []*Timer
	for rows.Next() {
		timer := &Timer{}
		err := rows.Scan(&timer.InternalID, &timer.ID, &timer.Message, &timer.User, &timer.Channel, &timer.Created, &timer.Due, &timer.SnoozedDue, &timer.SnoozeCount, &timer.Shown)
		if err != nil {
			return nil, err
		}
		timers = append(timers, timer)
	}
	return timers, nil
}

func markTimerAsShown(id string) error {
	_, err := db.Exec("UPDATE timers SET shown = true WHERE id = ?", id)
	return err
}

func newTimerID() (string, error) {
	for {
		id := randomString(4)

		_, err := getTimerByID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return id, nil // ID is unique
			}
			return "", err // Other database error
		}
		// If no error, timer with this ID already exists, loop again
	}
}

const letters = "abcdefghijklmnopqrstuvwxyz"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
