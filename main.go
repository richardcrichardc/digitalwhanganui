package main

import (
	"digitalwhanganui/validate"
	"github.com/richardcrichardc/binding"
	"github.com/richardcrichardc/martini"
	"github.com/richardcrichardc/render"
	"html"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var auckland *time.Location

func main() {
	// Decide where to load templated from
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

	// Martini Classic with parametric publicDir
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery(emailErrorMsg))
	m.Use(martini.Static(publicDir))
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

	rendererOptions := render.Options{Layout: "base", Directory: templateDir, Funcs: templateFuncs}
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

func para(text string) template.HTML {
	text = html.EscapeString(text)
	text = strings.Replace(text, "\r", "", -1)
	text = strings.Replace(text, "\n\n", "</p><p>", -1)
	text = strings.Replace(text, "\n", "<br>", -1)
	return template.HTML(text)
}

func formatTime(t time.Time) string {
	return t.In(auckland).Format("2-Feb-2006 3:04pm")
}

func obfEmail(email, label string) template.HTML {
	// Assumes strings are ascii
	emailBytes := []byte(email)
	for i := 0; i < len(emailBytes); i++ {
		emailBytes[i] = emailBytes[i] - 1
	}

	obfEmail := string(emailBytes)
	return template.HTML("<a class=\"obf-email\" href=\"#\" data-obf-email=\"" +
		html.EscapeString(obfEmail) + "\" data-obf-email-label=\"" +
		html.EscapeString(label) + "\">&nbsp;</a>")
}

func siteURL() string {
	if martini.Env == martini.Dev {
		return "http://localhost:3000"
	} else {
		return "http://xyzzy.digitalwhanganui.org.nz"
	}
}

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

	d.Submission, _ = s.Values["addme"].(ListingSubmission)
	r.HTML(200, "addme", d)
}

func addMePreview(r render.Render, s *Session) {
	var d listingPage

	submission, ok := s.Values["addme"].(ListingSubmission)

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

	sessionListingSubmission, _ := s.Values["addme"].(ListingSubmission)

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
	s.Values["addme"] = submission

	switch submission.Submit {
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
					"LoginLink": loginLink(shortAdministratorEmail())}
				sendMail(administratorEmail(), "New Listing", "newsubmission.tmpl", args)

				args["LoginLink"] = loginLink(submission.Listing.AdminEmail)
				sendMail(submission.Listing.FullAdminEmail(), "Digital Whanganui Submission", "pending.tmpl", args)

				s.Values["addMeDoneNewListing"] = true
				http.Redirect(w, req, "/addmedone", 302)
			case StatusAccepted:
				s.Values["addMeDoneNewListing"] = false
				http.Redirect(w, req, "/addmedone", 302)
			default:
				panic("Bad Status")
			}

			delete(s.Values, "addme")

		}
	case "edit":
		http.Redirect(w, req, "/addme", 302)
	default:
		r.StatusText(400, "Bad Request - Submit")
	}

}

func addMeDone(r render.Render, s *Session) {
	var d page

	newListing, ok := s.Values["addMeDoneNewListing"]
	if !ok {
		r.StatusText(400, "No Submission")
		return
	}

	if newListing.(bool) {
		d.Title = "Listing Submitted"
		d.Section = "addme"
		r.HTML(200, "addmedone", d)
	} else {
		d.Title = "Listing Updated"
		d.Section = "addme"
		r.HTML(200, "updatedone", d)
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

func uploadImage(r render.Render, req *http.Request) {
	//r.Status(500)
	//return

	var d struct {
		Id    string
		Error string
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		panic(err)
	}

	d.Id, d.Error = addImage(file)

	r.JSON(200, d)
}

func downloadImage(r render.Render, w http.ResponseWriter, req *http.Request, params martini.Params) {
	stream, created := fetchImage(params["imageId"], params["size"])
	if stream == nil {
		r.Status(404)
		return
	}

	http.ServeContent(w, req, "", *created, stream)
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

func login(r render.Render, params martini.Params, s *Session, w http.ResponseWriter, req *http.Request) {
	email := loginCodeToEmail(params["code"])

	if email == "" {
		invalidLoginCode(r, params["code"] == "lost")
		return
	}

	if email == shortAdministratorEmail() {
		s.Values["review"] = true
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
	s.Values["addme"] = submission
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

	if email == shortAdministratorEmail() || listingIdForAdminEmail(email) != 0 {
		args := map[string]string{"LoginLink": loginLink(email)}
		sendMail(email, "Digital Whanganui Login Link", "newloginlink.tmpl", args)

		d.Title = "New Login Link"
		r.HTML(200, "new-login-code", d)
	} else {
		d.Title = "New Login Link Failed"
		r.HTML(200, "new-login-code-failed", d)
	}
}

func reviewList(r render.Render, params martini.Params, s *Session) {
	if s.Values["review"] == nil {
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
	if s.Values["review"] == nil {
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
	if s.Values["review"] == nil {
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
