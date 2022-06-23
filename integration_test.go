package imageupsizer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEverything(t *testing.T) {
	var originalImage, err = GetImageConfigFromFile("./test.jpg")
	assert.NoError(t, err)

	largerImage, err := GetLargerImageFromFile("./test.jpg", "./")
	assert.NoError(t, err)

	assert.True(t, largerImage.Area > originalImage.Area)
	assert.True(t, largerImage.FileSize > originalImage.FileSize)

	assert.NoError(t, os.Remove(largerImage.LocalPath))
}
