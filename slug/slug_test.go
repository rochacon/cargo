package slug

import "testing"

func TestCleanURL(t *testing.T) {
	u := cleanURL("http://play.golang.org/somethingonthepath?one=1&two=2#damn")
	if u != "http://play.golang.org/somethingonthepath" {
		t.Fatal("Fail to cleanURL:", u)
	}
}
