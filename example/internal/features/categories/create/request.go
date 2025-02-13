package create

import "errors"

type Request struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *Request) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	return nil
}

var ErrEmptyName = error(errors.New("name cannot be empty"))
