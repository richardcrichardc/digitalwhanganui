package main

import (
	"html/template"
	"os"
	"time"

	"github.com/martini-contrib/binding"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var auckland *time.Location

func main() {
	loadConfig()
	initEmail()
	initDB()

	// Check if the flag file rescaleImages exists, if it does rescale all images
	// Images that have just been uploaded but the listing not saved will not be rescaled, so it is
	// advisable to resize twice a day or so apart
	if fileExists("rescaleImages") {
		go rescaleImages()
	}

	// Start goroutine that deletes images that have been uploaded but are not used
	go imageGC()

	// Martini Classic with parametric publicDir
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery(emailErrorMsg))
	m.Use(martini.Static(Config.PublicDir))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	// Sessions in database
	m.Use(SQLiteSession)

	// Renderer Options
	templateFuncs := []template.FuncMap{template.FuncMap{
		"para":       para,
		"shortDesc":  shortDesc,
		"formatTime": formatTime,
		"obfEmail":   obfEmail,
		"siteURL":    siteURL}}

	rendererOptions := render.Options{Layout: "base", Directory: Config.TemplateDir, Funcs: templateFuncs}
	m.Use(render.Renderer(rendererOptions))

	// Timezone for date formatting
	var err error
	auckland, err = time.LoadLocation("Pacific/Auckland")
	if err != nil {
		panic(err)
	}

	// Routing table
	r.Get("/", browse)
	r.Get("/browse/:majorMajorCat", browseCategory)
	r.Get("/browse/:majorMajorCat/:majorCat/:minorCat", browseCategorySummaries)
	r.Get("/browse/:majorMajorCat/:majorCat/:minorCat/:listingId", browseListing)
	r.Get("/individuals/", individuals)
	r.Get("/organisations/", organisations)
	r.Get("/individuals/:listingId", individualListing)
	r.Get("/organisations/:listingId", organisationListing)
	r.Get("/listing/:listingId", canonicalListing)
	r.Get("/addme", addMe)
	r.Post("/addme", binding.Bind(ListingSubmission{}), postAddMe)
	r.Post("/uploadImage", uploadImage)
	r.Get("/image/:imageId/:size", downloadImage)
	r.Get("/addmedone", addMeDone)
	r.Get("/faq", faq)
	r.Get("/about", about)
	r.Get("/search/", search)
	r.Get("/search/:listingId", searchListing)
	r.Get("/login/:code", login)
	r.Post("/login/:code", postLogin)
	r.Get("/review/", reviewList)
	r.Get("/review/:listingId", review)
	r.Post("/review/:listingId", postReview)
	r.Get("/export", export)
	r.Get("/500", fiveHundred)

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
