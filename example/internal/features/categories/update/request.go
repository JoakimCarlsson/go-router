package update

import "errors"

type Request struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (r *Request) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

var ErrEmptyName = error(errors.New("name cannot be empty"))
