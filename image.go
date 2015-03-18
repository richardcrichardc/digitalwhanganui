package main

import (
	"bytes"
	"database/sql"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nfnt/resize"
	"github.com/richardcrichardc/digitalwhanganui/martini"
	"github.com/richardcrichardc/digitalwhanganui/render"
)

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

	// read file into byte array
	var orig_buf bytes.Buffer
	_, err = orig_buf.ReadFrom(file)
	if err != nil {
		panic(err)
	}
	file.Close()
	orig := orig_buf.Bytes()

	d.Id, d.Error = addImage(orig)

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

func addImage(orig []byte) (string, string) {
	orig_image, format, err := image.Decode(bytes.NewReader(orig))
	if err != nil {
		return "", "Image could not be understood. Please upload a PNG, GIF or JPEG."
	}

	orig_size := orig_image.Bounds().Size()

	if orig_size.X < 692 && orig_size.Y < 606 {
		return "", "Image is too small. Please upload an image that is at least 692 pixels wide or 606 pixels high."
	}

	small := encode(resize.Thumbnail(255, 223, orig_image, resize.Lanczos3), format)
	large := encode(resize.Thumbnail(692, 606, orig_image, resize.Lanczos3), format)

	id := randomIdString()

	_, err = DB.Exec("INSERT INTO image(id, created, original, small, large) VALUES(?,datetime('now'),?,?,?)", id, orig, small, large)
	if err != nil {
		panic(err)
	}

	return id, ""
}

func encode(image image.Image, orig_format string) []byte {
	var buf bytes.Buffer
	var err error

	switch orig_format {
	case "jpeg":
		err = jpeg.Encode(&buf, image, &jpeg.Options{90})
	case "gif", "png":
		err = png.Encode(&buf, image)
	default:
		panic("Unsupported image format: " + orig_format)
	}
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func fetchImage(imageId, size string) (io.ReadSeeker, *time.Time) {
	// validate size string - very important to prevent SQL injection
	switch size {
	case "original":
	case "small":
	case "large":
		break
	default:
		return nil, nil
	}

	var imageBuf []byte
	var created time.Time

	row := DB.QueryRow("SELECT "+size+", created FROM image WHERE id=?", imageId)
	err := row.Scan(&imageBuf, &created)

	switch err {
	case nil:
		return bytes.NewReader(imageBuf), &created
	case sql.ErrNoRows:
		return nil, nil
	default:
		panic(err)
	}
}

func rescaleImages() {
	// New images will be created with new Ids. We don't want to rescale the images we have just
	// created. So get all the Ids to process first.

	var ids []string

	rows, err := DB.Query("SELECT id FROM image")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			panic(err)
		}

		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Create new image from each existing, update references
	// We want the new image to have a new id so that the old image does not stay cached in browsers
	// etc. However we cannot update references for listings the user is currently creating/editing

	for _, id := range ids {
		log.Println("Resizing image", id)
		// fetch original
		var orig []byte
		row := DB.QueryRow(`SELECT original FROM image WHERE id = ?`, id)
		err := row.Scan(&orig)
		if err != nil {
			panic(err)
		}

		// create new
		newId, imgErr := addImage(orig)
		if imgErr != "" {
			panic(imgErr)
		}

		// update references
		_, err = DB.Exec(`UPDATE listing SET ImageId=? WHERE ImageId = ?`, newId, id)
		if err != nil {
			panic(err)
		}

	}
}

func imageGC() {
	// Every midnight (NZT) delete any images from database that are no longer in use, that is, images
	// not referenced from the listing table. Images may have just been uploaded with the listing not
	// yet saved, so don't delete images less than a day old.

	NZT, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		panic(err)
	}

	for {
		now := time.Now().In(NZT)
		tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, NZT).Add(24 * time.Hour)
		duration := tomorrow.Sub(now)

		log.Println("Next ImageGC in", duration)
		time.Sleep(duration)

		log.Println("ImageGC start")
		_, err := DB.Exec(`DELETE FROM image WHERE id IN (
                        SELECT image.id
                        FROM image LEFT JOIN listing ON image.id=listing.imageId
                        WHERE listing.id IS NULL AND image.created < datetime('now', '-1 days')
                       )`)
		if err != nil {
			panic(err)
		}

		// Vacuum the database to recover free space
		// This locks the database for a few seconds
		_, err = DB.Exec(`VACUUM`)
		if err != nil {
			panic(err)
		}
		log.Println("ImageGC finish")
	}
}
