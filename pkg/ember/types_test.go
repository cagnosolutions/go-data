package ember

import (
	"fmt"
	"testing"
)

func TestStringValue(t *testing.T) {

	sv := stringValue{"foo bar"}

	fmt.Println(sv)

	out := sv.encode()
	fmt.Println(out)

	sv.decode()

}
