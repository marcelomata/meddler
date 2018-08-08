package meddler

import (
	"io"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
)

func TestLoad(t *testing.T) {
	once.Do(setup)
	insertAliceBob(t)

	elt := new(Person)
	elt.Age = 50
	elt.Closed = time.Now()
	if err := Load(db, "person", elt, 2); err != nil {
		t.Errorf("Load error on Bob: %v", err)
		return
	}
	bob.ID = 2
	personEqual(t, elt, bob)
	db.Exec("delete from person")
}

func TestLoadUint(t *testing.T) {
	once.Do(setup)
	insertAliceBob(t)

	elt := new(UintPerson)
	elt.Age = 50
	elt.Closed = time.Now()
	if err := Load(db, "person", elt, 2); err != nil {
		t.Errorf("Load error on Bob: %v", err)
		return
	}
	bob.ID = 2
	db.Exec("delete from person")
}

func TestSave(t *testing.T) {
	once.Do(setup)
	insertAliceBob(t)

	h := 73
	chris := &Person{
		ID:        0,
		Name:      "Chris",
		Email:     "chris@chris.com",
		Ephemeral: 19,
		Age:       23,
		Opened:    when.Local(),
		Closed:    when,
		Updated:   nil,
		Height:    &h,
	}

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("DB error on begin: %v", err)
	}
	if err = Save(tx, "person", chris); err != nil {
		t.Errorf("DB error on Save: %v", err)
	}

	id := chris.ID
	if id != 3 {
		t.Errorf("DB error on Save: expected ID of 3 but got %d", id)
	}

	chris.Email = "chris@chrischris.com"
	chris.Age = 27

	if err = Save(tx, "person", chris); err != nil {
		t.Errorf("DB error on Save: %v", err)
	}
	if chris.ID != id {
		t.Errorf("ID mismatch: found %d when %d expected", chris.ID, id)
	}
	if err = tx.Commit(); err != nil {
		t.Errorf("Commit error: %v", err)
	}

	// now test if the data looks right
	rows, err := db.Query("select * from person where id = ?", id)
	if err != nil {
		t.Errorf("DB error on query: %v", err)
		return
	}

	p := new(Person)
	if err = Default.ScanRow(rows, p); err != nil {
		t.Errorf("ScanRow error on Chris: %v", err)
		return
	}

	personEqual(t, p, &Person{3, "Chris", 0, "chris@chrischris.com", 0, 27, when, when, nil, &h})

	// delete this record so we don't confuse other tests
	if _, err = db.Exec("delete from person where id = ?", id); err != nil {
		t.Errorf("DB error on delete: %v", err)
	}
	db.Exec("delete from person")
}

func TestDriverErr(t *testing.T) {
	err, ok := DriverErr(io.EOF)
	if ok {
		t.Errorf("io.EOF: want driver error = false, got true")
	}
	if err != io.EOF {
		t.Errorf("io.EOF: want itself as returned error, got %v", err)
	}

	once.Do(setup)
	// insert into an invalid table
	alice.ID = 0
	err = Insert(db, "invalid", alice)
	if err == nil {
		t.Fatal("insert into invalid table, want error, got none")
	}
	err, ok = DriverErr(err)
	if !ok {
		t.Errorf("DriverErr: want ok to be true, got false")
	}
	if _, ok := err.(sqlite3.Error); !ok {
		t.Errorf("DriverErr: want sqlite3 error, got %T", err)
	}

	// insert with primary key set
	alice.ID = 1
	err = Insert(db, "person", alice)
	if err == nil {
		t.Errorf("insert with primary key already set. want error, got none")
	}

	// update with primary key not set
	alice.ID = 0
	err = Update(db, "person", alice)
	if err == nil {
		t.Errorf("update with primary key 0. want error, got none")
	}
}
