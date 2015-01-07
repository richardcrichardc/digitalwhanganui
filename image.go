package main

import (
	"bytes"
	"database/sql"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"time"
)

func addImage(file multipart.File) (string, string) {
	var orig_buf bytes.Buffer

	_, err := orig_buf.ReadFrom(file)
	if err != nil {
		panic(err)
	}
	file.Close()
	orig := orig_buf.Bytes()

	orig_image, format, err := image.Decode(bytes.NewReader(orig))
	if err != nil {
		return "", "Image could not be understood. Please upload a PNG, GIF or JPEG."
	}

	orig_size := orig_image.Bounds().Size()

	if orig_size.X < 692 && orig_size.Y < 606 {
		return "", "Image is too small. Please upload an image that is at least 692 pixels wide or 606 pixels high."
	}

	small := encode(resize.Thumbnail(255, 223, orig_image, resize.Lanczos3), format)
	large := encode(resize.Resize(692, 606, orig_image, resize.Lanczos3), format)

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
