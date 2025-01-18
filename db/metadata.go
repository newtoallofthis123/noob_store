package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/newtoallofthis123/noob_store/types"
)

// InsertMetaData inserts a metadata struct into the metadata table
func (db *Store) InsertMetaData(meta types.Metadata) error {
	_, err := db.pq.Insert("metadata").Columns("id", "name", "parent", "mime", "path", "user_id", "blob").
		Values(meta.Id, meta.Name, meta.Parent, meta.Mime, meta.Path, meta.UserId, meta.Blob).RunWith(db.db).Exec()

	return err
}

// GetMetaData gets the metadata by the name and path
func (db *Store) GetMetaDataByPath(path string) (types.Metadata, error) {
	row := db.pq.Select("*").From("metadata").Where("path LIKE ?", path).RunWith(db.db).QueryRow()

	var meta types.Metadata

	err := row.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
	if err != nil {
		return types.Metadata{}, err
	}

	return meta, nil
}

// GetMetaDataByDir gets the files and metadatas by the dir path
func (db *Store) GetMetaDataByDir(path string) ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").Where("parent LIKE ?", path).RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}

	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

// GetAllFiles gets all of the files in metadata
func (db *Store) GetAllFiles() ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}

	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

// GetMetaDataById gets the metadata by metadataId
func (db *Store) GetMetaDataById(id string) (types.Metadata, error) {
	row := db.pq.Select("*").From("metadata").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()

	var meta types.Metadata

	err := row.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
	if err != nil {
		return types.Metadata{}, err
	}

	return meta, nil
}

// GetMetadatasByUser gets all metadatas associated with a user
func (db *Store) GetMetadatasByUser(userId string) ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").Where(squirrel.Eq{"user_id": userId}).RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}
	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

// GetMetadataDirByUser gets all metadatas associated with a user under a dir
func (db *Store) GetMetadataDirByUser(userId, dir string) ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").Where("user_id LIKE ? AND parent LIKE ?", userId, dir).RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}
	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.UserId, &meta.CreatedAt)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

func (db *Store) DeleteMetadataById(id string) error {
	_, err := db.pq.Delete("metadata").Where(squirrel.Eq{"id": id}).RunWith(db.db).Exec()
	return err
}
