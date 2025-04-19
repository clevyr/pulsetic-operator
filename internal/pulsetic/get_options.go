package pulsetic

type FindRequest struct {
	URL *string
	ID  *int64
}

func (r FindRequest) Matches(m Monitor) bool {
	return (r.URL != nil && *r.URL == m.URL) || (r.ID != nil && *r.ID == m.ID)
}

type FindOption func(*FindRequest)

func FindByURL(url string) FindOption {
	return func(r *FindRequest) {
		r.URL = &url
	}
}

func FindByID(id int64) FindOption {
	return func(r *FindRequest) {
		r.ID = &id
	}
}
