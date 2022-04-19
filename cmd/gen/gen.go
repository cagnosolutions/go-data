//go:generate go run ../gen/cmd/modelgen.go
package gen

type Model interface {
	Select(string) error
	Insert(string) error
	Delete(string) error
}

type model struct {
	ID   int    `data:"id"`
	Name string `data:"name"`
}

func (m *model) Select(s string) error {
	return nil
}

func (m *model) Insert(s string) error {
	return nil
}

func (m *model) Delete(s string) error {
	return nil
}
