package livesync

import (
	"github.com/pkg/errors"
	cs "github.com/stratumn/go-chainscript"
)

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go go-connector/services/livesync Synchronizer

// Synchronizer is the type exposed by the livesync service.
type Synchronizer interface {
	Register() <-chan []*cs.Link
}

type synchronizer struct {
	registeredServices []<-chan []*cs.Link
}

func (s *synchronizer) Register() <-chan []*cs.Link {
	newCh := make(<-chan []*cs.Link)
	s.registeredServices = append(s.registeredServices, newCh)
	return nil
}

func (s *synchronizer) pollAndNotify() error {
	return errors.New("method is not implemented")
}
