package gensrc

//go:generate go run cmd/generator.go gensrc.User

type User struct {
	ID   int
	Name string
}

func String() string {
	return "Foo Bar"
}
