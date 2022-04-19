package gensrc

//go:generate go run cmd/main.go arg1 arg2

func PackFunc() string {
	return "cmd/gensrc/gensrc.PackFunc"
}
