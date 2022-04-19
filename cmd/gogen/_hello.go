package main

// this is a comment

// MyType is a type
type MyType struct {
	ID    int    `my-tag:"id"`
	Thing string `my-tag:"thing"`
}

func (m *MyType) String() string {
	return "Hello, world!"
}

func main() {
	t := new(MyType)
	fmt.Println(t)
}
