package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/newtoallofthis123/noob_store/types"
)

func (db *Store) InsertBlob(blob *types.Blob) error {
	_, err := db.pq.Insert("blobs").Columns("id", "name", "bucket", "size", "checksum", "start").Values(
		blob.Id, blob.Name, blob.Bucket, blob.Size, blob.Checksum, blob.Start).RunWith(db.db).Exec()
	return err
}

func (db *Store) GetBlob(name string) (types.Blob, error) {
	row := db.pq.Select("*").From("blobs").Where("name LIKE ?", name).RunWith(db.db).QueryRow()
	var blob types.Blob

	err := row.Scan(&blob.Id, &blob.Name, &blob.Bucket, &blob.Start, &blob.Size, &blob.Checksum, &blob.Deleted, &blob.Created_at)
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}

func (db *Store) GetBlobById(id string) (types.Blob, error) {
	row := db.pq.Select("*").From("blobs").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()
	var blob types.Blob

	err := row.Scan(&blob.Id, &blob.Name, &blob.Bucket, &blob.Start, &blob.Size, &blob.Checksum, &blob.Deleted, &blob.Created_at)
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}

func (db *Store) InsertMetaData(meta types.Metadata) error {
	_, err := db.pq.Insert("metadata").Columns("id", "name", "parent", "mime", "path", "blob").
		Values(meta.Id, meta.Name, meta.Parent, meta.Mime, meta.Path, meta.Blob).RunWith(db.db).Exec()

	return err
}

func (db *Store) GetMetaData(name, path string) (types.Metadata, error) {
	row := db.pq.Select("*").From("metadata").Where("name LIKE ? AND path LIKE ?", name, path).RunWith(db.db).QueryRow()

	var meta types.Metadata

	err := row.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.Created_at)
	if err != nil {
		return types.Metadata{}, err
	}

	return meta, nil
}

func (db *Store) GetMetaDataByDir(path string) ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").Where("parent LIKE ?", path).RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}

	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.Created_at)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

func (db *Store) GetAllFiles() ([]types.Metadata, error) {
	rows, err := db.pq.Select("*").From("metadata").RunWith(db.db).Query()
	if err != nil {
		return nil, err
	}

	metas := make([]types.Metadata, 0)

	for rows.Next() {

		var meta types.Metadata

		err := rows.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.Created_at)
		if err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

func (db *Store) GetMetaDataById(id string) (types.Metadata, error) {
	row := db.pq.Select("*").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()

	var meta types.Metadata

	err := row.Scan(&meta.Id, &meta.Name, &meta.Parent, &meta.Mime, &meta.Path, &meta.Blob, &meta.Created_at)
	if err != nil {
		return types.Metadata{}, err
	}

	return meta, nil
}
