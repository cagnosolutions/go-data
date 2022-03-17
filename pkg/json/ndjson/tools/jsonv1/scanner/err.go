package scanner

type OpErr struct {
	Pos     int
	Msg     string
	Content string
}

func (o OpErr) Error() string {
	return o.Msg + "; ..." + o.Content
}
