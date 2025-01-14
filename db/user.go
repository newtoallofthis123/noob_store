package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/newtoallofthis123/noob_store/types"
)

// CreateUser inserts a user
func (db *Store) CreateUser(user types.User) error {
	_, err := db.pq.Insert("users").Columns("id", "email", "password").Values(user.Id, user.Email, user.Password).RunWith(db.db).Exec()

	return err
}

// GetUser gets a user from the table using the user id
func (db *Store) GetUser(id string) (types.User, error) {
	row := db.pq.Select("*").From("users").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()

	var user types.User

	err := row.Scan(&user.Id, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return types.User{}, err
	}

	return user, nil
}

// GetUserByEmail gets a user from an email
func (db *Store) GetUserByEmail(email string) (types.User, error) {
	row := db.pq.Select("*").From("users").Where(squirrel.Eq{"email": email}).RunWith(db.db).QueryRow()

	var user types.User

	err := row.Scan(&user.Id, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return types.User{}, err
	}

	return user, nil
}

// CreateSession inserts a session
func (db *Store) CreateSession(session types.Session) error {
	_, err := db.pq.Insert("sessions").Columns("id", "user_id").Values(session.Id, session.UserId).RunWith(db.db).Exec()

	return err
}

// GetSession gets a session from the sessionId
func (db *Store) GetSession(id string) (types.Session, bool) {
	row := db.pq.Select("*").From("sessions").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()

	var session types.Session

	err := row.Scan(&session.Id, &session.UserId, &session.CreatedAt)
	if err != nil {
		return types.Session{}, false
	}

	return session, true
}

// TODO: Implement DeleteUser and DeleteSession
