package main

import (
	"github.com/martini-contrib/render"
	"github.com/richardcrichardc/digitalwhanganui/validate"
	"html/template"
	"net/http"
	"strings"
)

func addMe(r render.Render, s *Session, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		panic(err)
	}

	_, preview := req.Form["preview"]

	if preview {
		addMePreview(r, s)
	} else {
		addMeForm(r, s)
	}
}

func addMeForm(r render.Render, s *Session) {
	var d struct {
		page
		Cats       Categories
		Submission ListingSubmission
	}

	d.Title = "Add Me"
	d.Section = "addme"
	d.Cats = Cats
	d.JSFiles = []string{
		"/jquery.ui.widget.js",
		"/jquery.iframe-transport.js",
		"/jquery.fileupload.js",
		"/addme.js"}

	d.Submission, _ = s.Get("addme").(ListingSubmission)
	r.HTML(200, "addme", d)
}

func addMePreview(r render.Render, s *Session) {
	var d listingPage

	submission, ok := s.Get("addme").(ListingSubmission)

	if !ok {
		r.Redirect("/addme", 302)
		return
	}

	d.Title = "Add Me"
	d.Section = "addme"
	d.Listing = &submission.Listing
	d.Preview = true

	r.HTML(200, "browse-listing", d)
}

func postAddMe(r render.Render, formSubmission ListingSubmission, s *Session, w http.ResponseWriter, req *http.Request) {
	var submission ListingSubmission

	sessionListingSubmission, _ := s.Get("addme").(ListingSubmission)

	if formSubmission.FromPreview != "" {
		// Button pushed on Preview Page
		// Load form contents from session
		submission = sessionListingSubmission
		// Except which button was pushed
		submission.Submit = formSubmission.Submit
	} else {
		// Actual form submission
		submission = formSubmission
		// Retain listing Id from session in case Editing
		submission.Listing.Id = sessionListingSubmission.Listing.Id
		submission.Listing.Status = sessionListingSubmission.Listing.Status
	}

	if !submission.Listing.IsOrg {
		submission.Listing.Name = submission.Listing.AdminFirstName + " " + submission.Listing.AdminLastName
	}

	// coerce email address to lowercase
	submission.Listing.AdminEmail = strings.ToLower(submission.Listing.AdminEmail)

	errors := make(map[string]interface{})
	validate.Required(submission.Listing.AdminFirstName, "AdminFirstName", "First Name", errors)
	validate.Required(submission.Listing.AdminLastName, "AdminLastName", "Last Name", errors)
	validate.Required(submission.Listing.AdminPhone, "AdminPhone", "Telephone", errors)
	validate.Email(submission.Listing.AdminEmail, "AdminEmail", "Email", errors)
	if errors["AdminEmail"] == nil {
		adminEmailListingId := listingIdForAdminEmail(submission.Listing.AdminEmail)
		if adminEmailListingId != 0 &&
			submission.Listing.Id.Valid &&
			adminEmailListingId != int(submission.Listing.Id.Int64) {
			errors["AdminEmail"] = template.HTML(`There is another listing registered under that email address.
            If that is your email address you can <a href="/login/lost">login</a> and edit the other listing.`)
		}
	}
	validate.Required(submission.Listing.Name, "Name", "Name", errors)
	validate.Required(submission.Listing.Desc1, "Desc1", "Service / Product Description", errors)
	validate.Required(submission.Listing.Desc2, "Desc2", "About - Biography / Philosophy", errors)

	if submission.Listing.Phone == "" &&
		submission.Listing.Mobile == "" &&
		submission.Listing.Email == "" &&
		submission.Listing.Websites == "" &&
		submission.Listing.Address == "" {
		errors["Contact"] = "At least one contact method to publish must be provided."
	}

	submission.Listing.CatIds = parseCategories(submission.Categories)

	if len(submission.Listing.CatIds) == 0 {
		errors["Category"] = "At least one Category must be added."
	}

	submission.Errors = errors
	s.Set("addme", submission)

	switch submission.Submit {
	case "remove":
		removeListing(&submission.Listing)
		s.Set("addMeDoneStatus", StatusRemoved)
		http.Redirect(w, req, "/addmedone", 302)
		s.Delete("addme")
	case "preview":
		if len(errors) > 0 {
			http.Redirect(w, req, "/addme", 302)
		} else {
			http.Redirect(w, req, "/addme?preview", 302)
		}
	case "save":
		r.StatusText(500, "NOT IMPLEMENTED")
	case "submit":
		if len(errors) > 0 {
			http.Redirect(w, req, "/addme", 302)
		} else {
			storeListing(&submission.Listing)

			switch submission.Listing.Status {
			case StatusNew:
				args := map[string]string{
					"FirstName": submission.Listing.AdminFirstName,
					"Name":      submission.Listing.Name,
					"LoginLink": loginLink(Config.AdminEmailAddress)}
				sendMail(administratorEmail(), "New Listing", "newsubmission.tmpl", args)

				args["LoginLink"] = loginLink(submission.Listing.AdminEmail)
				sendMail(submission.Listing.FullAdminEmail(), "Digital Whanganui Submission", "pending.tmpl", args)

				http.Redirect(w, req, "/addmedone", 302)
			case StatusAccepted:
				http.Redirect(w, req, "/addmedone", 302)
			default:
				panic("Bad Status")
			}

			s.Set("addMeDoneStatus", submission.Listing.Status)
			s.Delete("addme")

		}
	case "edit":
		http.Redirect(w, req, "/addme", 302)
	default:
		r.StatusText(400, "Bad Request - Submit")
	}

}

func addMeDone(r render.Render, s *Session) {
	var d page

	listingStatus, ok := s.GetOK("addMeDoneStatus")
	if !ok {
		r.StatusText(400, "No Submission")
		return
	}

	switch listingStatus {
	case StatusNew:
		d.Title = "Listing Submitted"
		d.Section = "addme"
		r.HTML(200, "updatedone", d)
	case StatusAccepted:
		d.Title = "Listing Updated"
		d.Section = "addme"
		r.HTML(200, "updatedone", d)
	case StatusRemoved:
		d.Title = "Listing Removed"
		d.Section = "addme"
		r.HTML(200, "removedone", d)
	}
}

func parseCategories(categories string) (result []CategoryId) {
	if categories == "" {
		// No categories
		return
	}

	for _, cat := range strings.Split(categories, ",") {
		cat2 := strings.Split(cat, ".")
		if len(cat2) != 3 {
			panic("Bad category: " + cat)
		}
		result = append(result, CategoryId{cat2[0], cat2[1], cat2[2]})
	}
	return
}

func shortDesc(in string) string {
	splitPoint := strings.IndexRune(in, '\n')

	// First paragraph can be kept intact
	if splitPoint > -1 && splitPoint <= 300 {
		return in[0:splitPoint]
	}

	// Entire input fits in short desc and is not split
	if len(in) < 300 {
		return in
	}

	// Split after 300 chars and add elipsis
	return in[0:300] + "…"

}
