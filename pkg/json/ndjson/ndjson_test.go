package ndjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
)

type address struct {
	Name   string `json:"name"`
	Street string `json:"street"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
}

type user struct {
	ID       int                    `json:"id"`
	Name     []string               `json:"name"`
	Email    string                 `json:"email"`
	Addr     address                `json:"address"`
	IsActive bool                   `json:"is_active"`
	Notes    map[string]interface{} `json:"notes"`
}

var someNotes = []map[string]interface{}{
	{"march list": "nothing!!"},
	{"just a todo list": []string{"go to work\n", "pick up milk\n", "be ready to leave by 5\n"}},
	{"march todo": []string{"Work on redefining bad habits", "TikTok and chill"}},
	{"may todo": []string{"Play some video games", "Learn something new"}},
	{"april todo": "Spend some time alone"},
	{"asap": []string{"Have meals ready on time", "Work on taxes"}},
	{"when i have time": "Learn something new"},
	{"tonight": "Wash the dishes leftover"},
	{"every thursday": []string{"Work on redefining bad habits", "Family movie night", "Run computer backups"}},
	{"every other week": []string{"Volunteer work", "Family movie night"}},
	{"whenever i get time": []string{"Learn something new", "Pick up RX", "Work on taxes"}},
	{"june, sweden": "Pack for vacation"},
	{"friday night": []string{"Shoot some photography", "Play some video games"}},
	{"every day": []string{"Take a walk on the trail", "Family movie night"}},
	{"never": []string{"Clean room", "TikTok and chill"}},
	{"todo asap": "Prepare a proposal for work"},
}

var names = [][]string{
	{"Linette", "Wombwell"},
	{"Rosetta", "Flatte"},
	{"Cherie", "McGaffey"},
	{"Geralda", "Redit"},
	{"Delly", "Blench"},
	{"Erminie", "Tartt"},
	{"Cecil", "Bendix"},
	{"Pru", "Bispham"},
	{"Lauraine", "Jacob"},
	{"Edgar", "Dael"},
	{"Gianina", "Hardbattle"},
	{"Maryann", "Grubey"},
	{"Demetra", "Garroway"},
	{"Eduardo", "Tellesson"},
	{"Stanley", "Wraighte"},
	{"Thibaut", "Pennrington"},
	{"Vasily", "Presnall"},
	{"Glynis", "Stiller"},
	{"Irvin", "Gellibrand"},
	{"Clemmy", "Tripett"},
	{"Diandra", "Mosson"},
}

var addresses = []address{
	{Name: "Home Addr", Street: "51 Pennsylvania Center", City: "Trailsway", State: "AL", Zip: "35405"},
	{Name: "Home Addr", Street: "628 Oneill Avenue", City: "Warner", State: "AL", Zip: "35254"},
	{Name: "Home Addr", Street: "78333 Westridge Way", City: "Lerdahl", State: "AL", Zip: "35905"},
	{Name: "Home Addr", Street: "059 Evergreen Junction", City: "Fairview", State: "AL", Zip: "36119"},
	{Name: "Home Addr", Street: "33757 Monument Park", City: "Upham", State: "AL", Zip: "36605"},
	{Name: "Home Addr", Street: "628 Oneill Avenue", City: "Park", State: "AL", Zip: "35295"},
	{Name: "Home Addr", Street: "63809 Kipling Street", City: "Anniversary", State: "AL", Zip: "36134"},
	{Name: "Home Addr", Street: "9 Sachs Way", City: "Morning", State: "AL", Zip: "36119"},
	{Name: "Home Addr", Street: "0 Morning Drive", City: "Sullivan", State: "AL", Zip: "35205"},
	{Name: "Home Addr", Street: "805 Aberg Avenue", City: "Independence", State: "AL", Zip: "35263"},
	{Name: "Home Addr", Street: "3 Corben Street", City: "Gallagher", State: "AL", Zip: "36605"},
	{Name: "Home Addr", Street: "038 Hudson Crossing", City: "Bunting", State: "AL", Zip: "35254"},
	{Name: "Home Addr", Street: "32610 Northwestern Junction", City: "Macpherson", State: "AL", Zip: "35231"},
	{Name: "Home Addr", Street: "97960 Anzinger Alley", City: "Chinook", State: "AL", Zip: "35244"},
	{Name: "Home Addr", Street: "7485 Algoma Park", City: "6th", State: "AL", Zip: "36628"},
	{Name: "Home Addr", Street: "06025 Arrowood Avenue", City: "Hoard", State: "AL", Zip: "36689"},
	{Name: "Home Addr", Street: "51 Pennsylvania Center", City: "Warrior", State: "AL", Zip: "35244"},
	{Name: "Home Addr", Street: "63809 Kipling Street", City: "Morrow", State: "AL", Zip: "36195"},
	{Name: "Home Addr", Street: "0 Morning Drive", City: "Ash", State: "AL", Zip: "36114"},
	{Name: "Home Addr", Street: "038 Hudson Crossing", City: "Hayes", State: "AL", Zip: "36689"},
	{Name: "Home Addr", Street: "97960 Anzinger Alley", City: "Redwing", State: "AL", Zip: "35244"},
}

func makeEmail(name []string) string {
	s := fmt.Sprintf("%c%s@gmail.com", name[0][0], name[1])
	return strings.ToLower(s)
}

func populateUsersData() []user {
	var users []user
	for i := 0; i < 20; i++ {
		users = append(
			users, user{
				ID:       i + 1,
				Name:     names[i],
				Email:    makeEmail(names[i]),
				Addr:     addresses[i],
				IsActive: i%2 == 0,
				Notes:    someNotes[i%len(someNotes)],
			},
		)
	}
	return users
}

var file = new(bytes.Buffer)

func TestUserData(t *testing.T) {
	users := populateUsersData()
	for i, user := range users {
		b, err := json.MarshalIndent(user, "", "\t")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("user[%d]=%s\n", i, b)
	}
}

func writeDataWithLineWriter(w io.Writer, users []user, printFileContents bool) {

	// get our line writer
	lw := NewLineWriter(w)

	// range data
	for _, user := range users {
		// write each one to the file
		_, err := lw.Write(user)
		if err != nil {
			log.Panicf("got error=%s\n", err.Error())
		}
	}
	if printFileContents {
		fmt.Printf("file contents:\n%s\n", file.Bytes())
	}
}

func readDataWithLineReader(data []byte, printReadDocument bool) {

	// get our line reader
	lr := NewLineReader(bytes.NewReader(data))

	// declare our output
	var u user
	// read an entry
	err := lr.Read(&u)
	if err != nil {
		log.Panicf("got error=%s\n", err.Error())
	}
	if printReadDocument {
		fmt.Printf("record: %v\n", u)
	}
	// read rest of entries
	for i := 0; ; i++ {
		err = lr.Read(&u)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Panicf("got error=%s\n", err.Error())
		}
		if printReadDocument {
			fmt.Printf("record[%d]: %v\n", i, u)
		}
	}
}

func TestNewLineWriterWrite(t *testing.T) {

	// populate our user data
	users := populateUsersData()

	// test new line writer
	writeDataWithLineWriter(file, users, true)

}

func TestNewLineReaderRead(t *testing.T) {

	// reset our file just in case
	file.Reset()

	// populate our user data
	users := populateUsersData()

	// populate our file with data
	writeDataWithLineWriter(file, users, false)

	// test our line reader
	readDataWithLineReader(file.Bytes(), true)
}

func TestNewLineReaderReadAll(t *testing.T) {

	// reset our file just in case
	file.Reset()

	// populate our user data
	users := populateUsersData()

	// populate our file with data
	writeDataWithLineWriter(file, users, false)

	// test our line reader
	lr := NewLineReader(bytes.NewReader(file.Bytes()))

	// declare our output var
	var uu []user
	// print output
	fmt.Printf("[before] count=%d\n[before] output=%v\n", len(uu), uu)
	// read all
	_, err := lr.ReadAll(&uu)
	if err != nil {
		t.Errorf("got error=%s\n", err.Error())
	}
	// print output
	fmt.Printf("[after] count=%d\n", len(uu))
	for _, u := range uu {
		fmt.Printf("[after] output=%v\n", u)
	}
}

//
// func TestNewLineReaderReadOne(t *testing.T) {
//
// 	// reset our file just in case
// 	file.Reset()
//
// 	// populate our user data
// 	users := populateUsersData()
//
// 	// populate our file with data
// 	writeDataWithLineWriter(file, users, false)
//
// 	// test our line reader
// 	lr := NewLineReader(bytes.NewReader(file.Bytes()))
//
// 	// get our line reader
// 	lr := NewLineReader(bytes.NewReader(data))
//
// 	// declare our output
// 	var u user
// 	// read an entry
// 	err := lr.Read(&u)
// 	if err != nil {
// 		log.Panicf("got error=%s\n", err.Error())
// 	}
// 	if printReadDocument {
// 		fmt.Printf("record: %v\n", u)
// 	}
// 	// read rest of entries
// 	for i := 0; ; i++ {
// 		err = lr.Read(&u)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Panicf("got error=%s\n", err.Error())
// 		}
// 		if printReadDocument {
// 			fmt.Printf("record[%d]: %v\n", i, u)
// 		}
// 	}
// }
