package zoom

import (
	"time"
)

type Address struct {
	ID      int    `db:"id,pk"`
	City    string `db:"city"`
	Company string `db:"company"`
	Contact string `db:"contact"`
	Phone   string `db:"phone"`
	State   string `db:"state"`
	Street  string `db:"street"`
	UserID  int    `db:"user_id"`
	Zip     string `db:"zip"`
}

type User struct {
	ID                int       `db:"id,pk"`
	Active            bool      `db:"active,"`
	Username          string    `db:"username"`
	Password          string    `db:"password"`
	LastSeen          time.Time `db:"last_seen"`
	Created           time.Time `db:"created"`
	Role              string    `db:"role"`
	City              string    `db:"city"`
	Company           string    `db:"company"`
	Name              string    `db:"name"`
	Phone             string    `db:"phone"`
	State             string    `db:"state"`
	Street            string    `db:"street"`
	Zip               string    `db:"zip"`
	UsedCouponCodeIDS string    `json:"used_coupon_code_ids"`
}

type Stock struct {
	ID       int     `db:"id,pk"`
	Box      int     `db:"box"`
	Carton   int     `db:"carton"`
	Category string  `db:"category"`
	Desc1    string  `db:"desc_1"`
	Desc2    string  `db:"desc_2"`
	Envelope string  `db:"envelope"`
	Filter   string  `db:"filter"`
	Image    string  `db:"image"`
	Notes    string  `db:"notes"`
	Paper    string  `db:"paper"`
	Price    float32 `db:"price"`
	Weight   int     `db:"weight"`
	Restock  bool    `db:"restock"`
	Position string  `db:"position"`
	Template int     `db:"template"`
	Discount int     `db:"discount"`
}

type Quote struct {
	ID         int       `db:"id,pk"`
	Created    time.Time `db:"created"`
	PrintPrice float32   `db:"print_price"`
	Printing   string    `db:"printing"`
	Quantity   int       `db:"quantity"`
	ShipPrice  float32   `db:"ship_price"`
	Status     int       `db:"status"`
	Type       string    `db:"type"`
	UserID     int       `db:"user_id"`
	Vardat     int       `db:"vardat"`
	Zip        int       `db:"zip"`
	StockID    int       `db:"stock_id"`
	Portal     int       `db:"portal"`
	State      string    `db:"state"`
	Folded     int       `db:"folded"`
}
