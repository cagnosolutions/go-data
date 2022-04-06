package rest

type Pair struct {
	Key string
	Val interface{}
}

type ResourceItem interface {
	Pairs() []Pair
}

type ResourceCollection interface {

	// Get returns ResourceItem with matching id
	Get(id string) (ResourceItem, error)

	// Filter returns all ResourceItem that have
	// any supplied pars as matches
	Filter(by ...Pair) ([]ResourceItem, error)

	// Add adds a new ResourceItem using the sup-
	// plied id as an identifier and item as the
	// ResourceItem
	Add(id string, item ResourceItem) error

	// Set updates the ResourceItem containing
	// matching id using item as the updated
	// ResourceItem
	Set(id string, item interface{}) error

	// Del deletes the ResourceItem containing
	// the matching id
	Del(id string) error
}
