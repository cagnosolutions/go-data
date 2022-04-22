package items

type Item struct {
	ID   int    `tag:"pk"`
	Name string `tag:"text"`
}
