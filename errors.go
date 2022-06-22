package imageupsizer

import "errors"

var (
	errNoLargerAvailable = errors.New("there is no large image")
	errCaptcha           = errors.New("response was captcha page")
	errNoResults         = errors.New("no images found")
)
