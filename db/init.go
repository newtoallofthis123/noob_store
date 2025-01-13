package db

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
	pq squirrel.StatementBuilderType
}

func NewStore(connPath string) (*Store, error) {
	db, err := sql.Open("postgres", connPath)
	if err != nil {
		return nil, err
	}

	if db.Ping() != nil {
		return nil, err
	}

	return &Store{db: db, pq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}, nil
}

func (s *Store) InitTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS blobs(
		id text primary key,
		name text unique not null,
		bucket text not null,
		start bigint not null,
		size bigint not null,
		checksum bytea not null,
		deleted boolean default false,
		created_at timestamp default now()
	);

	CREATE TABLE IF NOT EXISTS metadata(
		id text primary key,
		name text not null,
		parent text not null,
		mime text not null,
		path text not null,
		blob text references blobs(id),
		created_at timestamp default now()
	);

	CREATE TABLE IF NOT EXISTS buckets(
		id text primary key,
		name text not null,
		path text not null,
		blobs text[]
	)
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
