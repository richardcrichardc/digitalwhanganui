package main

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func individualListing(r render.Render, params martini.Params) {
	displayListing(r, params, "Individuals")
}

func organisationListing(r render.Render, params martini.Params) {
	displayListing(r, params, "Organisations")
}

func searchListing(r render.Render, params martini.Params) {
	displayListing(r, params, "")
}

func canonicalListing(r render.Render, params martini.Params) {
	displayListing(r, params, "")
}

func displayListing(r render.Render, params martini.Params, majorMajorCatName string) {
	var d listingPage

	d.Section = "browse"

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
	d.MajorMajorCatName = majorMajorCatName
	d.MajorMajorCatURL = "."

	r.HTML(200, "browse-listing", d)
}

type Website struct {
	Font       string
	Icon       template.HTML
	URL, Label string
}

func (l *Listing) EachWebsite() (ret []Website) {
	for _, url := range strings.Split(strings.ToLower(l.Websites), "\n") {
		url = strings.TrimSpace(url)

		if url == "" {
			continue
		}

		if !strings.Contains(url, "://") {
			url = "http://" + url
		}

		var font, icon, label string
		if strings.Contains(url, "//www.facebook.com/") {
			font = "entypo-social"
			icon = "&#62221;"
			label = url
		} else if strings.Contains(url, "//twitter.com/") {
			font = "entypo-social"
			icon = "&#62230;"
			label = url
		} else {
			font = "entypo"
			icon = "&#127758;"
			label = url
		}

		ret = append(ret, Website{font, template.HTML(icon), url, label})
	}
	return
}
