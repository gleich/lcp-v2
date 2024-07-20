package strava

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/minio/minio-go/v7"
)

const bucketName = "mapbox-maps"

func FetchMap(polyline string) io.Reader {
	var (
		lineWidth = 2.0
		lineColor = "000"
		width     = 440
		height    = 240
		params    = url.Values{"access_token": {secrets.SECRETS.MapboxAccessToken}}
	)
	url := fmt.Sprintf(
		"https://api.mapbox.com/styles/v1/mattgleich/clxxsfdfm002401qj7jcxh47e/static/path-%f+%s(%s)/auto/%dx%d@2x?"+params.Encode(),
		lineWidth, lineColor, url.QueryEscape(polyline), width, height,
	)
	resp, err := http.Get(url)
	if err != nil {
		lumber.Error(err, "failed to fetch mapbox image with polyline", url)
		return nil
	}

	return resp.Body
}

func UploadMap(minioClient minio.Client, id uint64, reader io.Reader) {
	data, err := io.ReadAll(reader)
	if err != nil {
		lumber.Error(err, "reading from mapbox image failed")
	}

	_, err = minioClient.PutObject(
		context.Background(),
		bucketName,
		fmt.Sprintf("%d.png", id),
		reader,
		int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/png"},
	)
	if err != nil {
		lumber.Error(err, "failed to upload to minio")
	}
}

func RemoveOldMaps(minioClient minio.Client, activities []Activity) {
	var validKeys []string
	for _, activity := range activities {
		validKeys = append(validKeys, fmt.Sprintf("%d.png", activity.ID))
	}

	objects := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			lumber.Error(object.Err, "failed to load object")
			return
		}
		var validObject bool
		for _, key := range validKeys {
			if object.Key == key {
				validObject = true
			}
		}
		if !validObject {
			err := minioClient.RemoveObject(context.Background(), bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				lumber.Error(err, "failed to remove object")
				return
			}
		}
	}
}
