package cfgp

import (
	"os"
	"testing"
)

type myConf struct {
	Name    string
	Surname string `cfgp:"sur-key,specify the surname,"`
}

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-sur-key", "Doe", "-name", "Jonh"}
	c := myConf{}
	structValue, err := getStructValue(&c)
	if err != nil {
		t.Fatal(err)
	}
	err = parseFlags(structValue)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "Jonh" {
		t.Fatal("got:", c.Name, "expected: Jonh")
	}
	if c.Surname != "Doe" {
		t.Fatal("got:", c.Surname, "expected: Doe")
	}
}

func TestMakeHelpMessage(t *testing.T) {
	c := myConf{}
	structValue, err := getStructValue(&c)
	if err != nil {
		t.Fatal(err)
	}
	structType := structValue.Type()
	if f, ok := structType.FieldByName("Surname"); ok {
		m := makeHelpMessage(f)
		if m != "specify the surname, a string value" {
			t.Fatal("unexpected help message:", m)
		}
	} else {
		t.Fatal("parameter not found")
	}
}

func TestParse_invalid_format(t *testing.T) {
	Path = "local.yml"
	err := Parse(nil)
	if err == nil {
		t.Fail()
	}
}
