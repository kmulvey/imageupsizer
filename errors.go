package imageupsizer

import "errors"

var (
	ErrNoLargerAvailable = errors.New("there is no large image")
	ErrCaptcha           = errors.New("response was captcha page")
	ErrNoResults         = errors.New("no images found")
)
