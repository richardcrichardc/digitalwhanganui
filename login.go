package main

import (
	"net/http"
	"strings"

	"github.com/richardcrichardc/digitalwhanganui/martini"
	"github.com/richardcrichardc/digitalwhanganui/render"
)

func login(r render.Render, params martini.Params, s *Session, w http.ResponseWriter, req *http.Request) {
	email := loginCodeToEmail(params["code"])

	if email == "" {
		invalidLoginCode(r, params["code"] == "lost")
		return
	}

	if email == Config.AdminEmailAddress {
		s.Set("review", true)
		http.Redirect(w, req, "/review/", 302)
		return
	}

	listingId := listingIdForAdminEmail(email)
	if listingId == 0 {
		invalidLoginCode(r, false)
		return
	}

	var submission ListingSubmission
	submission.Listing = *fetchListing(listingId)
	s.Set("addme", submission)
	http.Redirect(w, req, "/addme", 302)
}

func invalidLoginCode(r render.Render, lost bool) {
	var d struct {
		page
		Lost bool
	}

	if lost {
		d.Title = "Lost Login Link"
	} else {
		d.Title = "Invalid Login Link"
	}

	d.Lost = lost
	r.HTML(200, "invalid-login-code", d)
}

func postLogin(r render.Render, req *http.Request) {
	var d page

	email := strings.TrimSpace(strings.ToLower(req.FormValue("email")))

	if email == Config.AdminEmailAddress || listingIdForAdminEmail(email) != 0 {
		args := map[string]string{"LoginLink": loginLink(email)}
		sendMail(email, "Digital Whanganui Login Link", "newloginlink.tmpl", args)

		d.Title = "New Login Link"
		r.HTML(200, "new-login-code", d)
	} else {
		d.Title = "New Login Link Failed"
		r.HTML(200, "new-login-code-failed", d)
	}
}
