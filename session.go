package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"html/template"
	"net/http"
	"time"

	"github.com/richardcrichardc/digitalwhanganui/martini"
)

type Session struct {
	Id      string
	Values  map[string]interface{}
	Expires time.Time
}

func SQLiteSession(c martini.Context, req *http.Request, res http.ResponseWriter) {
	session := Session{}

	// Get id from cookie and restore session from database
	cookie, err := req.Cookie("session")
	if err == nil {
		session.Id = cookie.Value
		session.Values = fetchSession(session.Id)
	}

	if session.Values == nil {
		// Session restore failed, create new session
		session.Id = randomIdString()
		session.Values = make(map[string]interface{})
	}

	// Update expiry
	session.Expires = time.Now().Add(time.Hour * 3)

	// Write session cookie
	http.SetCookie(res, &http.Cookie{
		Name:     "session",
		Value:    session.Id,
		Path:     "/",
		Expires:  session.Expires,
		HttpOnly: true})

	// Tell martini about session
	c.Map(&session)

	// Yield to next handler
	c.Next()

	// Save session to database
	storeSession(&session)
}

func randomIdString() string {
	id := make([]byte, 12)

	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(id)
}

func storeSession(session *Session) {
	// Marshall values to GOB
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(session.Values)
	if err != nil {
		panic(err)
	}

	_, err = DB.Exec("REPLACE INTO session(id, data, expires) VALUES(?,?,?)",
		session.Id,
		buf.Bytes(),
		session.Expires)

	if err != nil {
		panic(err)
	}
}

func fetchSession(sessionId string) map[string]interface{} {
	var data []byte

	row := DB.QueryRow("SELECT data FROM session WHERE id=?", sessionId)
	err := row.Scan(&data)

	switch err {
	case nil:
		break
	case sql.ErrNoRows:
		return nil
	default:
		panic(err)
	}

	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	var values map[string]interface{}
	err = dec.Decode(&values)
	if err != nil {
		panic(err)
	}

	return values
}

func init() {
	gob.Register(Listing{})
	gob.Register(ListingSubmission{})
	gob.Register(template.HTML(""))
}
