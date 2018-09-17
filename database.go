package main

import (
	"database/sql"
	"time"

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
	ListingCount int
	MinorCatKeys []string
	MinorCats    map[string]*MinorCat
}

type MinorCat struct {
	Name         string
	ListingCount int
}

type CategoryId struct {
	MajorMajorCode, MajorCode, MinorCode string
}

func (c *CategoryId) Name() (ret string) {
	majorMajorCat, ok := Cats.MajorMajorCats[c.MajorMajorCode]
	if !ok {
		return
	}
	//ret = majorMajorCat.Name

	majorCat, ok := majorMajorCat.MajorCats[c.MajorCode]
	if !ok {
		return
	}
	//ret = ret + " > " + majorCat.Name
	ret = majorCat.Name

	minorCat, ok := majorCat.MinorCats[c.MinorCode]
	if !ok {
		return
	}
	ret = ret + " > " + minorCat.Name
	return
}

func (c *CategoryId) URL() string {
	return "/browse/" + c.MajorMajorCode + "/" + c.MajorCode + "/" + c.MinorCode + "/"
}

type Listing struct {
	Id             sql.NullInt64
	Status         int
	AdminEmail     string `form:"adminEmail"`
	AdminFirstName string `form:"adminFirstName"`
	AdminLastName  string `form:"adminLastName"`
	AdminPhone     string `form:"adminPhone"`
	WCCExportOK    bool   `form:"WCCExportOK"`

	IsOrg    bool   `form:"isOrg"`
	Name     string `form:"name"`
	Desc1    string `form:"desc1"`
	Desc2    string `form:"desc2"`
	Phone    string `form:"phone"`
	Mobile   string `form:"mobile"`
	Email    string `form:"email"`
	Websites string `form:"websites"`
	Address  string `form:"address"`
	ImageId  string `form:"image"`

	CatIds []CategoryId
}

type ListingSubmission struct {
	Listing     Listing
	Categories  string `form:"categories"`
	Image       string `form:"image"`
	Submit      string `form:"submit"`
	FromPreview string `form:"fromPreview"`
	Errors      map[string]interface{}
}

type ListingSummary struct {
	Id        int
	Name      string
	ShortDesc string
	IsOrg     bool
	Sort      string
	ImageId   string
}

type ReviewListSummary struct {
	Id      int
	Name    string
	Updated time.Time
}

const (
	StatusNew      = 0
	StatusAccepted = 1
	StatusRejected = 2
	StatusExpired  = 3
	StatusRemoved  = 4
)

func storeListing(listing *Listing) {
	tx, err := DB.Begin()
	if err != nil {
		panic(err)
	}

	// Ensure delete triggers fire
	_, err = tx.Exec(`PRAGMA recursive_triggers = 1`)

	if err != nil {
		tx.Rollback()
		panic(err)
	}

	result, err := tx.Exec(`REPLACE INTO listing(id, status, adminEmail, adminFirstName, adminLastName, adminPhone, WCCExportOK, isOrg,
                        name, desc1, desc2, phone, email, websites, address, ImageId, updated) VALUES(?,?,lower(?),?,?,?,?,?,?,?,?,?,?,?,?,?, datetime('now'))`,
		listing.Id,
		listing.Status,
		listing.AdminEmail,
		listing.AdminFirstName,
		listing.AdminLastName,
		listing.AdminPhone,
		listing.WCCExportOK,
		listing.IsOrg,
		listing.Name,
		listing.Desc1,
		listing.Desc2,
		listing.Phone,
		listing.Email,
		listing.Websites,
		listing.Address,
		listing.ImageId)

	if err != nil {
		tx.Rollback()
		panic(err)
	}

	if !listing.Id.Valid {
		newListingId, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		listing.Id.Int64 = newListingId
		listing.Id.Valid = true
	}

	_, err = tx.Exec("DELETE FROM categoryListing WHERE listingId=?", listing.Id)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	stmt, err := tx.Prepare("INSERT INTO categoryListing(majorMajorCatCode, majorCatCode, minorCatCode, listingId) VALUES(?,?,?,?)")
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	for _, catId := range listing.CatIds {
		_, err = stmt.Exec(catId.MajorMajorCode, catId.MajorCode, catId.MinorCode, listing.Id)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	tx.Commit()
}

func removeListing(listing *Listing) {
	tx, err := DB.Begin()
	if err != nil {
		panic(err)
	}

	// Ensure delete triggers fire
	_, err = tx.Exec(`PRAGMA recursive_triggers = 1`)

	if err != nil {
		tx.Rollback()
		panic(err)
	}

	_, err = tx.Exec(`UPDATE listing SET status=4 WHERE id=?`, listing.Id)

	if err != nil {
		tx.Rollback()
		panic(err)
	}

	tx.Commit()
}

func fetchListing(listingId int) *Listing {
	var listing Listing
	row := DB.QueryRow(`SELECT id, status, adminEmail, adminFirstName, adminLastName, adminPhone, WCCExportOK, isOrg,
                        name, desc1, desc2, phone, email, websites, address, imageId FROM Listing WHERE id = ?`, listingId)

	err := row.Scan(
		&listing.Id,
		&listing.Status,
		&listing.AdminEmail,
		&listing.AdminFirstName,
		&listing.AdminLastName,
		&listing.AdminPhone,
		&listing.WCCExportOK,
		&listing.IsOrg,
		&listing.Name,
		&listing.Desc1,
		&listing.Desc2,
		&listing.Phone,
		&listing.Email,
		&listing.Websites,
		&listing.Address,
		&listing.ImageId)

	switch err {
	case nil:
		break
	case sql.ErrNoRows:
		return nil
	default:
		panic(err)
	}

	// Fetch Categories
	rows, err := DB.Query("SELECT majorMajorCatCode, majorCatCode, minorCatCode FROM categoryListing WHERE listingId=?", listingId)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var catId CategoryId
		if err := rows.Scan(&catId.MajorMajorCode, &catId.MajorCode, &catId.MinorCode); err != nil {
			panic(err)
		}
		listing.CatIds = append(listing.CatIds, catId)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	return &listing
}

func listingIdForAdminEmail(email string) int {
	var id int

	row := DB.QueryRow(`SELECT id FROM Listing WHERE adminEmail = lower(?)`, email)
	err := row.Scan(&id)

	switch err {
	case nil:
		return id
	case sql.ErrNoRows:
		return 0
	default:
		panic(err)
	}
}

func fetchCategorySummaries(majorMajorCatCode, majorCatCode, minorCatCode string) (summaries []ListingSummary) {
	if minorCatCode == "none" {
		rows, err := DB.Query("SELECT DISTINCT l.Id, l.Name, substr(desc1, 0, 320), l.isOrg, '', imageId FROM categoryListing cl JOIN listing l ON cl.listingId = l.id WHERE Status=1 AND majorMajorCatCode = ? AND majorCatCode = ? ORDER BY lower(Name)", majorMajorCatCode, majorCatCode)
		return fetchListingSummaries(rows, err)
	} else {
		rows, err := DB.Query("SELECT l.Id, l.Name, substr(desc1, 0, 320), l.isOrg, '', imageId FROM categoryListing cl JOIN listing l ON cl.listingId = l.id WHERE Status=1 AND majorMajorCatCode = ? AND majorCatCode = ? AND minorCatCode = ? ORDER BY lower(Name)", majorMajorCatCode, majorCatCode, minorCatCode)
		return fetchListingSummaries(rows, err)
	}
}

func fetchIndividualSummaries() (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT Id, Name, '', isOrg, upper(substr(adminLastName,1,1)), imageId FROM listing WHERE Status=1 AND isOrg=0 ORDER BY upper(adminLastName), upper(adminFirstName)")
	return fetchListingSummaries(rows, err)
}

func fetchOrganisationSummaries() (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT Id, Name, '', isOrg, upper(substr(Name,1,1)), imageId FROM listing WHERE Status=1 AND isOrg=1 ORDER BY upper(Name)")
	return fetchListingSummaries(rows, err)
}

func fetchSearchSummaries(search string) (summaries []ListingSummary) {
	rows, err := DB.Query("SELECT l.Id, l.Name, substr(l.desc1, 0, 320), l.isOrg, '', imageId FROM listing l JOIN listing_fts s ON l.id = s.docid WHERE l.Status=1 AND s.listing_fts MATCH ?", search)
	return fetchListingSummaries(rows, err)
}

func fetchListingSummaries(rows *sql.Rows, err error) (summaries []ListingSummary) {
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var summary ListingSummary
		if err := rows.Scan(&summary.Id, &summary.Name, &summary.ShortDesc, &summary.IsOrg, &summary.Sort, &summary.ImageId); err != nil {
			panic(err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func fetchReviewSummaries(status int) []ReviewListSummary {
	var summaries []ReviewListSummary

	rows, err := DB.Query("SELECT id, name, updated FROM listing WHERE status=? ORDER BY updated", status)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var summary ReviewListSummary
		if err := rows.Scan(&summary.Id, &summary.Name, &summary.Updated); err != nil {
			panic(err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	return summaries
}

func setListingStatus(listingId, newStatus int) {
	_, err := DB.Exec("UPDATE listing SET status=? WHERE id=?", newStatus, listingId)
	if err != nil {
		panic(err)
	}
}

func fetchExportData() (exportNotOKCount, exportOKCount int, exportData [][]string) {

	row := DB.QueryRow(`SELECT count(id) FROM Listing WHERE NOT WCCExportOK`)
	err := row.Scan(&exportNotOKCount)
	if err != nil {
		panic(err)
	}

	rows, err := DB.Query("SELECT id, adminEmail, adminFirstName, adminLastName, adminPhone, isOrg, name FROM listing WHERE status=1 AND WCCExportOK ORDER BY id")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id, adminEmail, adminFirstName, adminLastName, adminPhone, isOrg, name string

		if err := rows.Scan(&id, &adminEmail, &adminFirstName, &adminLastName, &adminPhone, &isOrg, &name); err != nil {
			panic(err)
		}
		exportData = append(exportData, []string{id, adminEmail, adminFirstName, adminLastName, adminPhone, isOrg, name})
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	exportOKCount = len(exportData)
	return
}

func updateCategoryCounts() {
	q := `SELECT
      mc.majormajorcatcode mmcc, mc.code mcc, NULL micc, count(l.id)
    FROM
      majorcat mc
      LEFT JOIN categoryListing cl ON mc.majormajorcatcode = cl.majormajorcatcode AND mc.code = cl.majorcatcode
      LEFT JOIN listing l ON cl.listingId = l.id AND l.status=1
    GROUP BY mcc, mcc, micc

    UNION

    SELECT
      mic.majormajorcatcode mmcc, mic.majorcatcode mcc, mic.code micc, count(l.id)
    FROM
      minorcat mic
      LEFT JOIN categoryListing cl ON mic.majormajorcatcode = cl.majormajorcatcode AND mic.majorcatcode = cl.majorcatcode AND mic.code = cl.minorcatcode
      LEFT JOIN listing l ON cl.listingId = l.id AND l.status=1
    GROUP BY mcc, mcc, micc;`

	rows, err := DB.Query(q)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var mmcc, mcc string
		var mic sql.NullString
		var count int

		if err := rows.Scan(&mmcc, &mcc, &mic, &count); err != nil {
			panic(err)
		}

		majorCat := Cats.MajorMajorCats[mmcc].MajorCats[mcc]
		if !mic.Valid {
			majorCat.ListingCount = count
		} else {
			minorCat := majorCat.MinorCats[mic.String]
			minorCat.ListingCount = count
		}

	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
}

func initDB() {
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
		majorMajorCat.MajorCats[code] = &MajorCat{name, 0, nil, make(map[string]*MinorCat)}
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
		majorCat.MinorCats[code] = &MinorCat{name, 0}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	updateCategoryCounts()

}

func loginLink(email string) string {
	code := randomIdString()
	_, err := DB.Exec(`INSERT INTO login(code, email, expires) VALUES(?,lower(?),datetime('now', '+365 days'))`, code, email)

	if err != nil {
		panic(err)
	}

	return siteURL() + "/login/" + code
}

func loginCodeToEmail(code string) string {
	var email string
	row := DB.QueryRow(`SELECT email FROM login WHERE code = ?`, code)
	err := row.Scan(&email)

	switch err {
	case nil:
		return email
	case sql.ErrNoRows:
		return ""
	default:
		panic(err)
	}
}
