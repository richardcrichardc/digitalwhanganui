package main

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/richardcrichardc/digitalwhanganui/martini"
	"github.com/richardcrichardc/digitalwhanganui/render"
)

func reviewList(r render.Render, params martini.Params, s *Session) {
	if s.Get("review") == nil {
		r.Status(403)
		return
	}

	var d struct {
		page
		NewListings      []ReviewListSummary
		AcceptedListings []ReviewListSummary
		RejectedListings []ReviewListSummary
		ExpiredListings  []ReviewListSummary
	}

	d.Title = "Review Listings"

	d.NewListings = fetchReviewSummaries(StatusNew)
	d.AcceptedListings = fetchReviewSummaries(StatusAccepted)
	d.RejectedListings = fetchReviewSummaries(StatusRejected)
	d.ExpiredListings = fetchReviewSummaries(StatusExpired)

	r.HTML(200, "reviewList", d)

}

func review(r render.Render, params martini.Params, s *Session) {
	if s.Get("review") == nil {
		r.Status(403)
		return
	}

	var d listingPage

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
	d.Review = true

	r.HTML(200, "browse-listing", d)
}

func postReview(r render.Render, params martini.Params, w http.ResponseWriter, req *http.Request, s *Session) {
	if s.Get("review") == nil {
		r.Status(403)
		return
	}

	listingIdString := params["listingId"]
	listingId, err := strconv.Atoi(listingIdString)
	if err != nil {
		r.Status(400)
		return
	}

	listing := fetchListing(listingId)
	if listing == nil {
		r.Status(404)
		return
	}

	action := req.FormValue("action")
	switch action {
	case "Accept":
		setListingStatus(listingId, StatusAccepted)
		updateCategoryCounts()
		args := map[string]string{
			"Id":          listingIdString,
			"FirstName":   listing.AdminFirstName,
			"Name":        listing.Name,
			"ListingLink": siteURL() + "/listing/" + listingIdString,
			"LoginLink":   loginLink(listing.AdminEmail)}
		sendMail(listing.FullAdminEmail(), "Digital Whanganui Submission Accepted", "accepted.tmpl", args)
	case "Reject":
		setListingStatus(listingId, StatusRejected)
		updateCategoryCounts()
	default:
		r.StatusText(400, "Bad action: "+action)
	}

	http.Redirect(w, req, "/review/"+listingIdString, 302)
}

func export(r render.Render, w http.ResponseWriter, req *http.Request, s *Session) {
	if s.Get("review") == nil {
		r.Status(403)
		return
	}

	exportNotOKCount, exportOKCount, data := fetchExportData()

	header := [][]string{
		{"DigitalWhanganui WRC Export"},
		{},
		{strconv.Itoa(exportOKCount), "Marked WRC Export OK"},
		{strconv.Itoa(exportNotOKCount), "Marked WRC Export Not OK"},
		{},
		{"ID", "Email", "First Name", "Last Name", "Phone", "Is Organisation", "Listing Name"},
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Content-Disposition", `attachment; filename="export.csv"`)

	cw := csv.NewWriter(w)
	cw.WriteAll(header)
	cw.WriteAll(data)
	cw.Flush()
}

func fiveHundred(r render.Render, w http.ResponseWriter, req *http.Request, s *Session) {
	panic("The sky is falling")
}
