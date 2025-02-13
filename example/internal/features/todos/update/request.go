package update

import "errors"

type Request struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

func (r *Request) Validate() error {
	if r.Title != nil && *r.Title == "" {
		return ErrEmptyTitle
	}
	return nil
}

var ErrEmptyTitle = error(errors.New("title cannot be empty"))
