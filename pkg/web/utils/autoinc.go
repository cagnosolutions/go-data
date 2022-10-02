package utils

import "sync"

type AutoID struct {
	sync.Mutex
	id int
}

func (a *AutoID) ID() (id int) {
	a.Lock()
	defer a.Unlock()
	id = a.id
	a.id++
	return
}
