package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var Cats Categories

type Categories struct {
	MajorCatKeys []string
	MajorCats    map[string]*MajorCat
}

type MajorCat struct {
	Name         string
	MinorCatKeys []string
	MinorCats    map[string]*MinorCat
}

type MinorCat struct {
	Name string
}

type CategoryId struct {
	MajorCode, MinorCode string
}

type Listing struct {
	Id             int
	AdminEmail     string `form:"adminEmail"`
	AdminFirstName string `form:"adminFirstName"`
	AdminLastName  string `form:"adminLastName"`
	AdminPhone     string `form:"adminPhone"`

	Name     string `form:"name"`
	Desc1    string `form:"desc1"`
	Desc2    string `form:"desc2"`
	Phone    string `form:"phone"`
	Mobile   string `form:"mobile"`
	Email    string `form:"email"`
	Websites string `form:"websites"`
	Address  string `form:"address"`
}

type ListingSubmission struct {
	Listing     Listing
	Categories  string `form:"categories"`
	Image       string `form:"image"`
	Submit      string `form:"submit"`
	FromPreview string `form:"fromPreview"`
	CatIds      []CategoryId
	Errors      map[string]string
}

type ListingSummary struct {
	Id   int
	Name string
}

func storeListing(submission ListingSubmission) {
	result, err := DB.Exec(`REPLACE INTO listing(adminEmail, adminFirstName, adminLastName, adminPhone,
                        name, desc1, desc2, phone, email, websites, address) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		submission.Listing.AdminEmail,
		submission.Listing.AdminFirstName,
		submission.Listing.AdminLastName,
		submission.Listing.AdminPhone,
		submission.Listing.Name,
		submission.Listing.Desc1,
		submission.Listing.Desc2,
		submission.Listing.Phone,
		submission.Listing.Email,
		submission.Listing.Websites,
		submission.Listing.Address)

	if err != nil {
		panic(err)
	}

	listingId, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	// TODO delete old categoruListings

	stmt, err := DB.Prepare("INSERT INTO categoryListing(majorCatCode, minorCatCode, listingId) VALUES(?,?,?)")
	if err != nil {
		panic(err)
	}

	for _, catId := range submission.CatIds {
		_, err = stmt.Exec(catId.MajorCode, catId.MinorCode, listingId)
		if err != nil {
			panic(err)
		}
	}
}

func fetchListing(listingId int) *Listing {
	var listing Listing
	row := DB.QueryRow(`SELECT id, adminEmail, adminFirstName, adminLastName, adminPhone,
                        name, desc1, desc2, phone, email, websites, address FROM Listing WHERE id = ?`, listingId)

	err := row.Scan(
		&listing.Id,
		&listing.AdminEmail,
		&listing.AdminFirstName,
		&listing.AdminLastName,
		&listing.AdminPhone,
		&listing.Name,
		&listing.Desc1,
		&listing.Desc2,
		&listing.Phone,
		&listing.Email,
		&listing.Websites,
		&listing.Address)

	switch err {
	case nil:
		return &listing
	case sql.ErrNoRows:
		return nil
	default:
		panic(err)
	}
}

func fetchListingSummaries(majorCatCode, minorCatCode string) (summaries []ListingSummary) {

	rows, err := DB.Query("SELECT l.Id, l.Name FROM categoryListing cl JOIN listing l ON cl.listingId = l.id WHERE majorCatCode = ? AND minorCatCode = ?", majorCatCode, minorCatCode)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var summary ListingSummary
		if err := rows.Scan(&summary.Id, &summary.Name); err != nil {
			panic(err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func init() {
	var err error
	DB, err = sql.Open("sqlite3", "./digitalwhanganui.sdb")
	if err != nil {
		panic(err)
	}

	// Load major categories
	Cats.MajorCats = make(map[string]*MajorCat)
	rows, err := DB.Query("SELECT code, name FROM majorCat ORDER BY sort")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			panic(err)
		}
		Cats.MajorCatKeys = append(Cats.MajorCatKeys, code)
		Cats.MajorCats[code] = &MajorCat{name, nil, make(map[string]*MinorCat)}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Load minor categories
	rows, err = DB.Query("SELECT majorCatCode, code, name FROM minorCat ORDER BY sort")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var majorCatCode, code, name string
		if err := rows.Scan(&majorCatCode, &code, &name); err != nil {
			panic(err)
		}

		majorCat, ok := Cats.MajorCats[majorCatCode]
		if !ok {
			panic("No majorCat: " + majorCatCode)
		}

		majorCat.MinorCatKeys = append(majorCat.MinorCatKeys, code)
		majorCat.MinorCats[code] = &MinorCat{name}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
}
