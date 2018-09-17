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

	"github.com/go-martini/martini"
)

type Session struct {
	Id      string
	Expires time.Time
	values  map[string]interface{}
	dirty   bool
}

func (s *Session) Get(key string) interface{} {
	return s.values[key]
}

func (s *Session) GetOK(key string) (interface{}, bool) {
	value, ok := s.values[key]
	return value, ok
}

func (s *Session) Set(key string, value interface{}) {
	s.values[key] = value
	s.dirty = true
}

func (s *Session) Delete(key string) {
	delete(s.values, key)
}

func SQLiteSession(c martini.Context, req *http.Request, res http.ResponseWriter) {
	session := Session{}

	// Get id from cookie and restore session from database
	cookie, err := req.Cookie("session")
	if err == nil {
		session.Id = cookie.Value
		session.values, session.Expires = fetchSession(session.Id)
	}

	if session.values == nil {
		// Session restore failed, create new session
		session.Id = randomIdString()
		session.values = make(map[string]interface{})
	}

	// Calculate expiry
	// Truncated to 10 minutes so we do not need to update the database for
	// every request
	// Convert to UTC so comparision with date read in from database works
	expires := time.Now().Add(time.Hour * 3).Truncate(time.Minute * 10).UTC()

	// Write session cookie
	http.SetCookie(res, &http.Cookie{
		Name:     "session",
		Value:    session.Id,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true})

	// Tell martini about session
	c.Map(&session)

	// Yield to next handler
	c.Next()

	// Save session to database when needed
	if session.dirty || expires != session.Expires {
		session.Expires = expires
		storeSession(&session)
	}
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
	err := enc.Encode(session.values)
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

func fetchSession(sessionId string) (map[string]interface{}, time.Time) {
	var data []byte
	var expiresString string

	row := DB.QueryRow("SELECT data, expires FROM session WHERE id=?", sessionId)
	err := row.Scan(&data, &expiresString)

	switch err {
	case nil:
		break
	case sql.ErrNoRows:
		return nil, time.Time{}
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

	expires, err := time.Parse("2006-01-02 15:04:05", expiresString)
	if err != nil {
		panic(err)
	}

	return values, expires
}

func init() {
	gob.Register(Listing{})
	gob.Register(ListingSubmission{})
	gob.Register(template.HTML(""))
}
