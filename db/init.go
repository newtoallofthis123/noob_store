package db

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

// Store represents the database and query builder interface
type Store struct {
	db *sql.DB
	pq squirrel.StatementBuilderType
}

// NewStore initializes a new Store instance and pings the database
func NewStore(connPath string) (Store, error) {
	db, err := sql.Open("postgres", connPath)
	if err != nil {
		return Store{}, err
	}

	return Store{db: db, pq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}, nil
}

// InitTables initialized tables if they are not already existing
func (s *Store) InitTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS blobs(
		id text primary key,
		name text unique not null,
		bucket text not null,
		start bigint not null,
		size bigint not null,
		checksum text not null,
		deleted boolean default false,
		created_at timestamp default now()
	);

	CREATE TABLE IF NOT EXISTS users(
		id text primary key,
		email text not null,
		password text not null,
		created_at timestamp default now()
	);

	CREATE TABLE IF NOT EXISTS sessions(
		id text primary key,
		user_id text references users(id),
		created_at timestamp default now()
	);

	CREATE TABLE IF NOT EXISTS metadata(
		id text primary key,
		name text not null,
		parent text not null,
		mime text not null,
		path text not null,
		blob text references blobs(id),
		user_id text references users(id),
		created_at timestamp default now()
	);
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
