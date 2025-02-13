package create

import "errors"

type Request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (r *Request) Validate() error {
	if r.Title == "" {
		return ErrEmptyTitle
	}
	return nil
}

var ErrEmptyTitle = error(errors.New("title cannot be empty"))
