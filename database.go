package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var Cats Categories

type Categories struct {
	MajorMajorCatKeys []string
	MajorMajorCats    map[string]*MajorMajorCat
}

type MajorMajorCat struct {
	Name         string
	Blurb        string
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
	MajorMajorCode, MajorCode, MinorCode string
}

type Listing struct {
	Id             int
	AdminEmail     string `form:"adminEmail"`
	AdminFirstName string `form:"adminFirstName"`
	AdminLastName  string `form:"adminLastName"`
	AdminPhone     string `form:"adminPhone"`

	IsOrg    bool   `form:"isOrg"`
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
	Sort string
}

func storeListing(submission ListingSubmission) {
	result, err := DB.Exec(`REPLACE INTO listing(adminEmail, adminFirstName, adminLastName, adminPhone, isOrg,
                        name, desc1, desc2, phone, email, websites, address) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		submission.Listing.AdminEmail,
		submission.Listing.AdminFirstName,
		submission.Listing.AdminLastName,
		submission.Listing.AdminPhone,
		submission.Listing.IsOrg,
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
	row := DB.QueryRow(`SELECT id, adminEmail, adminFirstName, adminLastName, adminPhone, isOrg,
                        name, desc1, desc2, phone, email, websites, address FROM Listing WHERE id = ?`, listingId)

	err := row.Scan(
		&listing.Id,
		&listing.AdminEmail,
		&listing.AdminFirstName,
		&listing.AdminLastName,
		&listing.AdminPhone,
		&listing.IsOrg,
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

func fetchCategorySummaries(majorCatCode, minorCatCode string) (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT l.Id, l.Name FROM categoryListing cl JOIN listing l ON cl.listingId = l.id WHERE majorCatCode = ? AND minorCatCode = ?", majorCatCode, minorCatCode)
	return fetchListingSummaries(rows, err)
}

func fetchIndividualSummaries() (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT Id, Name, upper(substr(adminLastName,1,1)) FROM listing WHERE isOrg=0 ORDER BY upper(adminLastName), upper(adminFirstName)")
	return fetchListingSummaries(rows, err)
}

func fetchOrganisationSummaries() (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT Id, Name, upper(substr(Name,1,1)) FROM listing WHERE isOrg=1 ORDER BY Name")
	return fetchListingSummaries(rows, err)
}

func fetchListingSummaries(rows *sql.Rows, err error) (summaries []ListingSummary) {

	for rows.Next() {
		var summary ListingSummary
		if err := rows.Scan(&summary.Id, &summary.Name, &summary.Sort); err != nil {
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

	// Load majormajor categories
	Cats.MajorMajorCats = make(map[string]*MajorMajorCat)
	rows, err := DB.Query("SELECT code, name, blurb FROM majorMajorCat ORDER BY sort")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var code, name, blurb string
		if err := rows.Scan(&code, &name, &blurb); err != nil {
			panic(err)
		}
		Cats.MajorMajorCatKeys = append(Cats.MajorMajorCatKeys, code)
		Cats.MajorMajorCats[code] = &MajorMajorCat{name, blurb, nil, make(map[string]*MajorCat)}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Load major categories
	rows, err = DB.Query("SELECT majorMajorCatCode, code, name FROM majorCat ORDER BY sort")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var majorMajorCatCode, code, name string
		if err := rows.Scan(&majorMajorCatCode, &code, &name); err != nil {
			panic(err)
		}

		majorMajorCat, ok := Cats.MajorMajorCats[majorMajorCatCode]
		if !ok {
			panic("No majorMajorCat: " + majorMajorCatCode)
		}

		majorMajorCat.MajorCatKeys = append(majorMajorCat.MajorCatKeys, code)
		majorMajorCat.MajorCats[code] = &MajorCat{name, nil, make(map[string]*MinorCat)}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Load minor categories
	rows, err = DB.Query("SELECT majorMajorCatCode, majorCatCode, code, name FROM minorCat ORDER BY sort")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var majorMajorCatCode, majorCatCode, code, name string
		if err := rows.Scan(&majorMajorCatCode, &majorCatCode, &code, &name); err != nil {
			panic(err)
		}

		majorMajorCat, ok := Cats.MajorMajorCats[majorMajorCatCode]
		if !ok {
			panic("No majorMajorCat: " + majorMajorCatCode)
		}

		majorCat, ok := majorMajorCat.MajorCats[majorCatCode]
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
