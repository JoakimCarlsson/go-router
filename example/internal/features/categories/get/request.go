package get

import "errors"

type Request struct {
	ID string `json:"id"`
}

func (r *Request) Validate() error {
	if r.ID == "" {
		return ErrInvalidID
	}
	return nil
}

var ErrInvalidID = error(errors.New("invalid category ID"))
