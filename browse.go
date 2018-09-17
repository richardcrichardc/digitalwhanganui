package main

import (
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type page struct {
	Title        string
	CanonicalURL string
	Section      string
	JSFiles      []string
	SearchQuery  string
}

type listingPage struct {
	page
	MajorMajorCatName string
	MajorMajorCatURL  string
	MajorCat          *MajorCat
	MinorCat          *MinorCat
	Listing           *Listing
	Preview           bool
	Review            bool
}

type summaryPage struct {
	page
	ListingSummaries []ListingSummary
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

func browseCategory(r render.Render, params martini.Params) {
	var d struct {
		page
		MajorMajorCat *MajorMajorCat
	}

	majorMajorCatCode := params["majorMajorCat"]
	majorMajorCat := Cats.MajorMajorCats[majorMajorCatCode]

	d.Title = majorMajorCat.Name
	d.Section = "browse"
	d.MajorMajorCat = majorMajorCat

	r.HTML(200, "browse-category", d)
}

func browseCategorySummaries(r render.Render, params martini.Params) {
	var d struct {
		page
		MajorMajorCat    *MajorMajorCat
		MajorCat         *MajorCat
		MinorCat         *MinorCat
		ListingSummaries []ListingSummary
	}

	majorMajorCatCode := params["majorMajorCat"]
	majorMajorCat := Cats.MajorMajorCats[majorMajorCatCode]
	majorCatCode := params["majorCat"]
	majorCat := majorMajorCat.MajorCats[majorCatCode]
	minorCatCode := params["minorCat"]
	minorCat := majorCat.MinorCats[minorCatCode]

	d.Title = majorCat.Name
	d.Section = "browse"
	d.MajorMajorCat = majorMajorCat
	d.MajorCat = majorCat
	d.MinorCat = minorCat
	d.ListingSummaries = fetchCategorySummaries(majorMajorCatCode, majorCatCode, minorCatCode)

	r.HTML(200, "browse-summaries", d)
}

func browseListing(r render.Render, params martini.Params) {
	var d listingPage

	majorMajorCatCode := params["majorMajorCat"]
	majorMajorCat := Cats.MajorMajorCats[majorMajorCatCode]
	majorCatCode := params["majorCat"]
	majorCat := majorMajorCat.MajorCats[majorCatCode]
	minorCatCode := params["minorCat"]
	minorCat := majorCat.MinorCats[minorCatCode]

	d.Section = "browse"
	d.MajorMajorCatName = majorMajorCat.Name
	d.MajorMajorCatURL = "../.."
	d.MajorCat = majorCat
	d.MinorCat = minorCat

	listingId, err := strconv.Atoi(params["listingId"])
	if err != nil {
		r.Status(400)
		return
	}

	d.Listing = fetchListing(listingId)
	if d.Listing == nil || d.Listing.Status != StatusAccepted {
		r.Status(404)
		return
	}

	d.Title = d.Listing.Name
	d.CanonicalURL = siteURL() + "/listing/" + params["listingId"]

	r.HTML(200, "browse-listing", d)
}

func individuals(r render.Render) {
	var d summaryPage
	d.Title = "Individuals"
	d.ListingSummaries = fetchIndividualSummaries()
	azSummaries(r, d)
}

func organisations(r render.Render) {
	var d summaryPage
	d.Title = "Organisations"
	d.ListingSummaries = fetchOrganisationSummaries()
	azSummaries(r, d)
}

func azSummaries(r render.Render, d summaryPage) {
	var lastSort string
	for i := range d.ListingSummaries {
		if d.ListingSummaries[i].Sort == lastSort {
			d.ListingSummaries[i].Sort = ""
		} else {
			lastSort = d.ListingSummaries[i].Sort
		}
	}

	r.HTML(200, "az-summaries", d)
}

func faq(r render.Render) {
	var d page
	d.Title = "Frequently Asked Questions"
	d.Section = "faq"
	r.HTML(200, "faq", d)
}

func about(r render.Render) {
	var d page
	d.Title = "About"
	d.Section = "about"
	r.HTML(200, "about", d)
}

func search(r render.Render, req *http.Request) {
	var d struct {
		page
		ListingSummaries []ListingSummary
	}

	d.Title = "Search"
	d.SearchQuery = req.FormValue("q")
	d.ListingSummaries = fetchSearchSummaries(d.SearchQuery)

	r.HTML(200, "search", d)
}
