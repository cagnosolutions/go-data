//go:generate go run modelgen.go
package _scratch

import (
	"fmt"
)

type FooModel struct {
	ID   int
	Desc string
}

func NewFooModel() *FooModel {
	return new(FooModel)
}

func (m *FooModel) GetID() int {
	return m.ID
}

func (m *FooModel) GetDesc() string {
	return m.Desc
}

func (m *FooModel) String() string {
	return fmt.Sprintf("%#v\n", m)
}
