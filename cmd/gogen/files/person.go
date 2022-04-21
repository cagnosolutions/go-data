package files

import (
	"time"
)

type Person struct {
	Name    string
	Age     int
	Created time.Time
}

// AUTO GENERATED SETTERS

func (p *Person) SetName(s string) {
	p.Name = s
}
func (p *Person) SetAge(i int) {
	p.Age = i
}
func (p *Person) SetCreated(t time.Time) {
	p.Created = t
}

// AUTO GENERATED GETTERS

func (p *Person) GetName() string {
	return p.Name
}
func (p *Person) GetAge() int {
	return p.Age
}
func (p *Person) GetCreated() time.Time {
	return p.Created
}
