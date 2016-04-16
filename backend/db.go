package backend

import (
	"database/sql"
)

type ElectionsDb struct {
	db *sql.DB
}

func (edb ElectionsDb) StartTransaction() (*ElectionsTx, error) {
	if tx, err := edb.db.Begin(); nil != err {
		return nil, err
	} else {
		return &ElectionsTx{tx: tx}, nil
	}
}

func ConnectDatabase(db *sql.DB) (ElectionsDb, error) {
	db.Exec(`PRAGMA foreign_keys = ON`)
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS user (
	uid INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT UNIQUE,
	token TEXT UNIQUE,
	siteadmin BOOLEAN NOT NULL DEFAULT 0
)
`); nil != err {
		return ElectionsDb{}, err
	}

	if _, err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS user_unique_unregistered ON user (name) WHERE email IS NULL;
`); nil != err {
		return ElectionsDb{}, err
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS election (
	eid INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	title TEXT NOT NULL DEFAULT '',
	candidates TEXT NOT NULL,
	closed BOOLEAN NOT NULL DEFAULT 0,
	public BOOLEAN NOT NULL DEFAULT 0,
	open BOOLEAN NOT NULL DEFAULT 0,
	editopen BOOLEAN NOT NULL DEFAULT 0
);
`); nil != err {
		return ElectionsDb{}, err
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS vote (
	eid INTEGER NOT NULL REFERENCES election ON DELETE CASCADE ON UPDATE CASCADE,
	uid INTEGER NOT NULL REFERENCES user ON DELETE RESTRICT ON UPDATE CASCADE,
	ranking TEXT NOT NULL,
	UNIQUE (eid, uid)
);
`); nil != err {
		return ElectionsDb{}, err
	}

	return ElectionsDb{
		db: db,
	}, nil
}
