package main

import (
	"digitalwhanganui/validate"
	"fmt"
	"github.com/richardcrichardc/binding"
	"github.com/richardcrichardc/martini"
	"github.com/richardcrichardc/render"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var templateDir, publicDir string

	if fileExists("templates") {
		templateDir = "templates"
	} else {
		templateDir = "/usr/local/share/digitalwhanganui/templates"
	}

	if fileExists("public") {
		publicDir = "public"
	} else {
		publicDir = "/usr/local/share/digitalwhanganui/public"
	}

	// Classic with parametric publicDir
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Use(martini.Static(publicDir))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	m.Use(SQLiteSession)

	rendererOptions := render.Options{Layout: "base", Directory: templateDir}
	m.Use(render.Renderer(rendererOptions))

	r.Get("/", browse)
	r.Get("/browse/:majorCat/", browseMajor)
	r.Get("/browse/:majorCat/:minorCat", browseMinor)
	r.Get("/browse/:majorCat/:minorCat/:listingId", browseListing)
	r.Get("/addme", addMe)
	r.Post("/addme", binding.Bind(ListingSubmission{}), postAddMe)
	r.Post("/uploadImage", uploadImage)
	r.Get("/about", about)
	r.Get("/search", search)

	m.Run()
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

type page struct {
	Title   string
	Section string
	JSFiles []string
}

func browse(r render.Render) {
	var d struct {
		page
		Cats Categories
		Name string
	}

	d.Title = "Digital Whanganui"
	d.Section = "browse"
	d.Cats = Cats
	d.Name = "Bob"

	r.HTML(200, "browse", d)
}

func browseMajor(r render.Render, params martini.Params) {
	var d struct {
		page
		MajorCat *MajorCat
	}

	majorCatCode := params["majorCat"]
	majorCat := Cats.MajorCats[majorCatCode]

	d.Title = majorCat.Name
	d.Section = "browse"
	d.MajorCat = majorCat

	r.HTML(200, "browse-major", d)
}

func browseMinor(r render.Render, params martini.Params) {
	var d struct {
		page
		MajorCat         *MajorCat
		MinorCat         *MinorCat
		ListingSummaries []ListingSummary
	}

	majorCatCode := params["majorCat"]
	majorCat := Cats.MajorCats[majorCatCode]
	minorCatCode := params["minorCat"]
	minorCat := majorCat.MinorCats[minorCatCode]

	d.Title = majorCat.Name
	d.Section = "browse"
	d.MajorCat = majorCat
	d.MinorCat = &minorCat
	d.ListingSummaries = fetchListingSummaries(majorCatCode, minorCatCode)

	r.HTML(200, "browse-minor", d)
}

func browseListing(r render.Render, params martini.Params) {
	var d struct {
		page
		MajorCat *MajorCat
		MinorCat *MinorCat
		Listing  *Listing
	}

	majorCatCode := params["majorCat"]
	majorCat := Cats.MajorCats[majorCatCode]
	minorCatCode := params["minorCat"]
	minorCat := majorCat.MinorCats[minorCatCode]

	d.Section = "browse"
	d.MajorCat = majorCat
	d.MinorCat = &minorCat

	listingId, err := strconv.Atoi(params["listingId"])
	if err != nil {
		r.Status(400)
		return
	}

	d.Listing = fetchListing(listingId)
	if d.Listing == nil {
		r.Status(404)
		return
	}

	d.Title = d.Listing.Name

	r.HTML(200, "browse-listing", d)
}

func addMe(r render.Render, s *Session) {
	var d struct {
		page
		Cats    Categories
		Listing ListingSubmission
	}

	d.Title = "Add Me"
	d.Section = "addme"
	d.Cats = Cats
	d.JSFiles = []string{
		"/jquery.ui.widget.js",
		"/jquery.iframe-transport.js",
		"/jquery.fileupload.js",
		"/addme.js"}

	d.Listing, _ = s.Values["addme"].(ListingSubmission)

	fmt.Println(d.Listing)

	r.HTML(200, "addme", d)
}

func postAddMe(r render.Render, listing ListingSubmission, s *Session, w http.ResponseWriter, req *http.Request) {
	errors := make(map[string]string)
	validate.Required(listing.AdminFirstName, "AdminFirstName", "First Name", errors)
	validate.Required(listing.AdminLastName, "AdminLastName", "Last Name", errors)
	validate.Required(listing.AdminPhone, "AdminPhone", "Telephone", errors)
	validate.Email(listing.AdminEmail, "AdminEmail", "Email", errors)
	validate.Required(listing.Name, "Name", "Name", errors)
	validate.Required(listing.Desc1, "Desc1", "Service / Product Description", errors)
	validate.Required(listing.Desc2, "Desc2", "About - Biography / Philosophy", errors)

	if listing.Phone == "" &&
		listing.Mobile == "" &&
		listing.Email == "" &&
		listing.Websites == "" &&
		listing.Address == "" {
		errors["Contact"] = "At least one contact method to publish must be provided."
	}

	listing.CatIds = parseCategories(listing.Categories)

	if len(listing.CatIds) == 0 {
		errors["Category"] = "At least one Category must be added."
	}

	switch listing.Submit {
	case "preview":
		r.StatusText(500, "NOT IMPLEMENTED")
	case "save":
		r.StatusText(500, "NOT IMPLEMENTED")
	case "submit":
		if len(errors) > 0 {
			listing.Errors = errors
			s.Values["addme"] = listing
			http.Redirect(w, req, "/addme", 302)
		} else {
			r.Status(200)
		}
		storeListing(listing)
	default:
		r.StatusText(400, "Bad Request - Submit")
	}

}

func parseCategories(categories string) (result []CategoryId) {
	if categories == "" {
		// No categories
		return
	}

	for _, cat := range strings.Split(categories, ",") {
		cat2 := strings.Split(cat, ".")
		if len(cat2) != 2 {
			panic("Bad category: " + cat)
		}
		result = append(result, CategoryId{cat2[0], cat2[1]})
	}
	return
}

func uploadImage(r render.Render) {
	var d struct {
		Id    string
		Error string
	}

	d.Id = "123456789"

	time.Sleep(2 * time.Second)
	r.JSON(200, d)
}

func about(r render.Render) {
	var d struct {
		page
		Cats Categories
		Name string
	}

	d.Title = "About"
	d.Section = "about"
	d.Cats = Cats
	d.Name = "Bob"

	r.HTML(200, "about", d)
}

func search(r render.Render) {
	var d struct {
		page
		Cats Categories
		Name string
	}

	d.Title = "Search"
	d.Cats = Cats
	d.Name = "Bob"

	r.HTML(200, "search", d)
}
