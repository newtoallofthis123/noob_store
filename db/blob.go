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

	err := row.Scan(&blob.Id, &blob.Name, &blob.Bucket, &blob.Start, &blob.Size, &blob.Checksum, &blob.Deleted, &blob.CreatedAt)
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}

func (db *Store) GetBlobById(id string) (types.Blob, error) {
	row := db.pq.Select("*").From("blobs").Where(squirrel.Eq{"id": id}).RunWith(db.db).QueryRow()
	var blob types.Blob

	err := row.Scan(&blob.Id, &blob.Name, &blob.Bucket, &blob.Start, &blob.Size, &blob.Checksum, &blob.Deleted, &blob.CreatedAt)
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}
