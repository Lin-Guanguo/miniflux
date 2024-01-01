// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"context"
	"database/sql"
	"time"

	"miniflux.app/logger"
)

// Storage handles all operations related to the database.
type Storage struct {
	db *DBWrapper
}

type DBWrapper struct {
	*sql.DB
}

func (db *DBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	logger.Debug("[Database] SQL:%s, args:%v", query, args)
	return db.DB.Exec(query, args...)
}

func (db *DBWrapper) Query(query string, args ...any) (*sql.Rows, error) {
	logger.Debug("[Database] SQL:%s, args:%v", query, args)
	return db.DB.Query(query, args...)
}

func (db *DBWrapper) QueryRow(query string, args ...any) *sql.Row {
	logger.Debug("[Database] SQL:%s, args:%v", query, args)
	return db.DB.QueryRow(query, args...)
}

// NewStorage returns a new Storage.
func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		&DBWrapper{db},
	}
}

// DatabaseVersion returns the version of the database which is in use.
func (s *Storage) DatabaseVersion() string {
	var dbVersion string
	err := s.db.QueryRow(`SELECT current_setting('server_version')`).Scan(&dbVersion)
	if err != nil {
		return err.Error()
	}

	return dbVersion
}

// Ping checks if the database connection works.
func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

// DBStats returns database statistics.
func (s *Storage) DBStats() sql.DBStats {
	return s.db.Stats()
}
