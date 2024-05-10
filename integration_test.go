package imageupsizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
func TestEverything(t *testing.T) {
	t.Parallel()

	var originalImage, err = GetImageConfigFromFile("./test.jpg")
	assert.NoError(t, err)

	largerImage, err := GetLargerImageFromFile("./test.jpg", "./")
	assert.NoError(t, err)

	assert.True(t, largerImage.Area > originalImage.Area)
	assert.True(t, largerImage.FileSize > originalImage.FileSize)
	assert.NoError(t, os.Remove(largerImage.LocalPath))
}
*/

func TestErrorImg(t *testing.T) {
	t.Parallel()

	var isError, err = isErrorImage(&ImageData{LocalPath: "./error-image.jpg"})
	assert.NoError(t, err)
	assert.True(t, isError)
}
