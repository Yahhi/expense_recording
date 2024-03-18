package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"ingresos_gastos/config"
	"ingresos_gastos/storage_interface"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PostgresAdapter struct {
	dbInside *sql.DB
}

func NewPostgresAdapter(cfg config.Config) PostgresAdapter {
	var err error
	var db *sql.DB
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=require",
		cfg.PGHost, cfg.PGPort, cfg.PGAdmin, cfg.PGPass, cfg.PGDbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return PostgresAdapter{db}
}

func (db PostgresAdapter) CreateUser(userID int64, name string) error {
	var id int
	err := db.dbInside.QueryRow("INSERT INTO users (id, name, created) VALUES ($1, $2, $3) RETURNING id", userID, name, time.Now()).Scan(&id)
	if err != nil {
		return fmt.Errorf("error creating User: %v", err)
	}
	return nil
}

func (db PostgresAdapter) UserExists(userID int64) (bool, error) {
	row := db.dbInside.QueryRow("SELECT name FROM users WHERE id = $1", userID)
	var name string
	err := row.Scan(&name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error requesting UserExists: %v", err)
	}
	return true, nil
}

func (db PostgresAdapter) SetState(userID int64, state string) error {
	_, err := db.dbInside.Exec("UPDATE users SET status = $1 WHERE id = $2", state, userID)
	if err != nil {
		return fmt.Errorf("error setting User state: %v", err)
	}
	return nil
}

func (db PostgresAdapter) GetUserState(userID int64) (string, error) {
	var state string
	err := db.dbInside.QueryRow("SELECT status FROM users WHERE id = $1", userID).Scan(&state)
	if err != nil {
		return state, fmt.Errorf("error selecting User state: %v", err)
	}
	return state, nil
}

func (db PostgresAdapter) CreateTarget(tag string, amount float32, periodStart, periodEnd time.Time, userID int64) error {
	_, errGettingExistingUser := db.dbInside.Exec("DELETE FROM targets WHERE tag = $1 AND period_start = $2 AND user_id = $3", tag, periodStart, userID)
	if errGettingExistingUser != nil {
		return fmt.Errorf("error deleting possibly existing previous target: %v", errGettingExistingUser)
	}
	_, err := db.dbInside.Exec("INSERT INTO targets (tag, amount, period_start, period_end, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id", tag, amount, periodStart, periodEnd, userID)
	if err != nil {
		return fmt.Errorf("error creating target for tag '%s' and user %d: %v", tag, userID, err)
	}
	return nil
}

func (db PostgresAdapter) GetTargets(periodStart time.Time, periodEnd time.Time, userID int64) ([]storage_interface.Target, error) {
	var targets []storage_interface.Target
	rows, err := db.dbInside.Query("SELECT id, tag, amount, period_start, period_end, user_id FROM targets WHERE user_id = $1 AND period_start = $2 AND period_end = $3 ", userID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("error selecting targets in GetTargets: %v", err)
	}

	for rows.Next() {
		var target storage_interface.Target
		if err := rows.Scan(&target.ID, &target.Tag, &target.Amount, &target.PeriodStart, &target.PeriodEnd, &target.UserID); err != nil {
			return nil, fmt.Errorf("error unwrapping targets in GetTargets: %v", err)
		}
		targets = append(targets, target)
	}

	return targets, nil
}

// CreateMoneyEvent creates a new money event in the database
func (db PostgresAdapter) CreateMoneyEvent(amount float32, currency, comment, tag string, userID int64) error {
	_, err := db.dbInside.Exec("INSERT INTO money_events (amount, currency, comment, tag, created, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", amount, currency, comment, tag, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("error creating movey event: %v", err)
	}
	return nil
}

func (db PostgresAdapter) GetMoneyEventsByDateInterval(startDate, endDate time.Time, userID int64) ([]storage_interface.MoneyEvent, error) {
	var events []storage_interface.MoneyEvent

	rows, err := db.dbInside.Query("SELECT id, amount, currency, comment, tag, created, user_id FROM money_events WHERE created >= $1 AND created <= $2 AND user_id = $3 ORDER BY created ", startDate, endDate, userID)
	if err != nil {
		return nil, fmt.Errorf("error selecting money events: %v", err)
	}

	for rows.Next() {
		var event storage_interface.MoneyEvent
		if err := rows.Scan(&event.ID, &event.Amount, &event.Currency, &event.Comment, &event.Tag, &event.Created, &event.UserID); err != nil {
			return nil, fmt.Errorf("error unwrapping money event in GetMoneyEventsByDateInterval: %v", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (db PostgresAdapter) AddTagForUser(tag string, userID int64) error {
	_, err := db.dbInside.Query("INSERT INTO accepted_tags (tag, user_id) VALUES ($1, $2) RETURNING id", tag, userID)
	if err != nil {
		return fmt.Errorf("error adding tag '%s' for user %d: %v", tag, userID, err)
	}
	return nil
}

func (db PostgresAdapter) RemoveTagForUser(tag string, userID int64) error {
	_, err := db.dbInside.Exec("DELETE FROM accepted_tags WHERE tag = $1 AND user_id = $2", tag, userID)
	if err != nil {
		return fmt.Errorf("error deleting tag '%s' for user %d: %v", tag, userID, err)
	}
	return nil
}

func (db PostgresAdapter) GetUserTags(userID int64) ([]string, error) {
	var tags []string

	rows, err := db.dbInside.Query("SELECT tag FROM accepted_tags WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error querring user tags: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("error unwrapping user tags in GetUserTags: %v", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (db PostgresAdapter) SaveFeedback(userID int64, message string) error {
	_, err := db.dbInside.Query("INSERT INTO feedback (message, user_id) VALUES ($1, $2) RETURNING id", message, userID)
	if err != nil {
		return fmt.Errorf("error saving feedback: %v", err)
	}
	return nil
}
