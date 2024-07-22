package strava

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"net/url"

	"github.com/buckket/go-blurhash"
	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
	"github.com/minio/minio-go/v7"
)

const bucketName = "mapbox-maps"

func FetchMap(polyline string) []byte {
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

	var b bytes.Buffer
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		lumber.Error(err, "failed to read data from request")
		return nil
	}

	return b.Bytes()
}

func MapBlurData(data []byte) *string {
	reader := bytes.NewReader(data)
	parsedPNG, err := png.Decode(reader)
	if err != nil {
		lumber.Error(err, "decoding PNG failed")
		return nil
	}

	width := parsedPNG.Bounds().Dx()
	height := parsedPNG.Bounds().Dy()
	blurData, err := blurhash.Encode(4, 3, parsedPNG)
	if err != nil {
		lumber.Error(err, "encoding png into blurhash failed")
		return nil
	}

	scaleDownFactor := 25
	blurImage, err := blurhash.Decode(blurData, width/scaleDownFactor, height/scaleDownFactor, 1)
	if err != nil {
		lumber.Error(err, "decoding blurhash data into img failed")
		return nil
	}
	blurImageBuffer := new(bytes.Buffer)
	err = png.Encode(blurImageBuffer, blurImage)
	if err != nil {
		lumber.Error(err, "creating png based off blurred image failed")
		return nil
	}
	blurDataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(blurImageBuffer.Bytes())
	return &blurDataURI
}

func UploadMap(minioClient minio.Client, id uint64, data []byte) {
	reader := bytes.NewReader(data)
	size := int64(len(data))

	_, err := minioClient.PutObject(
		context.Background(),
		bucketName,
		fmt.Sprintf("%d.png", id),
		reader,
		size,
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
