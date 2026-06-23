package upload

import "errors"

var (
	ErrMissingToken = errors.New("upload: missing ecto1 access token")
	ErrBadToken     = errors.New("upload: access token must start with ecto1:")
)
