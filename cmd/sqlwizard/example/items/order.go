package items

type Order struct {
	ID    int     `tag:"pk"`
	Items []Item  `tag:"o2m,fk"`
	Total float64 `tag:""`
}
