package displateApi

type Store interface {
	GetLimitedEditionDisplates() []Displate
}

type Database interface {
	StoreDisplates([]Displate) error
}

type Backend struct {
	displates []Displate
}

func (b *Backend) GetLimitedEditionDisplates() []Displate {
	//TODO implement me)
	return b.displates
}

func (b *Backend) StoreDisplates(displates []Displate) error {
	//TODO implement me
	b.displates = displates
	return nil
}

func NewBackend() *Backend {
	return &Backend{
		displates: make([]Displate, 0),
	}
}
