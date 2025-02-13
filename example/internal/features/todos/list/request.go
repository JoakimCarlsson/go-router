package list

type Request struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	Done   bool `json:"done"`
}

func (r *Request) Validate() error {
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}