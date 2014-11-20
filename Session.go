package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"github.com/richardcrichardc/martini"
	"net/http"
	"time"
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
	session.Expires = time.Now().Add(time.Hour * 4)

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
}

/*
package main

import (
    "bytes"
    "encoding/gob"
    "fmt"
    "log"
)

type P struct {
    X, Y, Z int
    Name    string
}

type Q struct {
    X, Y *int32
    Name string
}

// This example shows the basic usage of the package: Create an encoder,
// transmit some values, receive them with a decoder.
func main() {
    // Initialize the encoder and decoder.  Normally enc and dec would be
    // bound to network connections and the encoder and decoder would
    // run in different processes.
    var network bytes.Buffer        // Stand-in for a network connection
    enc := gob.NewEncoder(&amp;network) // Will write to network.
    dec := gob.NewDecoder(&amp;network) // Will read from network.

    // Encode (send) some values.
    err := enc.Encode(P{3, 4, 5, "Pythagoras"})
    if err != nil {
        log.Fatal("encode error:", err)
    }
    err = enc.Encode(P{1782, 1841, 1922, "Treehouse"})
    if err != nil {
        log.Fatal("encode error:", err)
    }

    // Decode (receive) and print the values.
    var q Q
    err = dec.Decode(&amp;q)
    if err != nil {
        log.Fatal("decode error 1:", err)
    }
    fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)
    err = dec.Decode(&amp;q)
    if err != nil {
        log.Fatal("decode error 2:", err)
    }
    fmt.Printf("%q: {%d, %d}\n", q.Name, *q.X, *q.Y)

}
*/
