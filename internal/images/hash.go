package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/buckket/go-blurhash"
	"github.com/gleich/lumber/v3"
)

func BlurImage(data []byte) []byte {
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
	return blurImageBuffer.Bytes()
}

func BlurDataURI(data []byte) string {
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(data))
}
