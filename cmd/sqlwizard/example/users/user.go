package users

type user struct {
	id         int `tag:"pk"`
	first_name string
	last_name  string
	full_name  string
	billing    address `tag:"fk"`
	shipping   address `tag:"fk"`
}

type address struct {
	id     int
	street [2]string
	city   string
	state  string
	zip    string
}
