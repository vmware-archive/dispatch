package prettyjson_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
)

func Example() {
	v := map[string]interface{}{
		"str":   "foo",
		"num":   100,
		"bool":  false,
		"null":  nil,
		"array": []string{"foo", "bar", "baz"},
		"map": map[string]interface{}{
			"foo": "bar",
		},
	}
	s, _ := prettyjson.Marshal(v)
	fmt.Println(string(s))
	// Output:
	// {
	//   [34;1m"array"[0m: [
	//     [32;1m"foo"[0m,
	//     [32;1m"bar"[0m,
	//     [32;1m"baz"[0m
	//   ],
	//   [34;1m"bool"[0m: [33;1mfalse[0m,
	//   [34;1m"map"[0m: {
	//     [34;1m"foo"[0m: [32;1m"bar"[0m
	//   },
	//   [34;1m"null"[0m: [30;1mnull[0m,
	//   [34;1m"num"[0m: [36;1m100[0m,
	//   [34;1m"str"[0m: [32;1m"foo"[0m
	// }
}

func TestMarshal(t *testing.T) {
	prettyJson := func(s string) string {
		var v interface{}

		err := json.Unmarshal([]byte(s), &v)

		if err != nil {
			t.Error(err)
		}

		prettyJsonByte, err := prettyjson.Marshal(v)

		if err != nil {
			t.Error(err)
		}

		return string(prettyJsonByte)
	}

	test := func(expected, actual string) {
		if expected != actual {
			t.Errorf("\nexpected:\n%s\n\nactual:\n%s", expected, actual)
		}
	}

	blueBold := color.New(color.FgBlue, color.Bold).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyanBold := color.New(color.FgCyan, color.Bold).SprintFunc()
	blackBold := color.New(color.FgBlack, color.Bold).SprintFunc()
	yelloBold := color.New(color.FgYellow, color.Bold).SprintFunc()

	actual := prettyJson(`{
  "key": {
    "a": "str",
    "b": 100,
    "c": null,
    "d": true,
    "e": false,
    "f": { "key": "str" },
	"g": {},
	"h": []
  }
}`)

	expectedFormat := `{
  %s: {
    %s: %s,
    %s: %s,
    %s: %s,
    %s: %s,
    %s: %s,
    %s: {
      %s: %s
    },
    %s: {},
    %s: []
  }
}`

	expected := fmt.Sprintf(expectedFormat,
		blueBold(`"key"`),
		blueBold(`"a"`), greenBold(`"str"`),
		blueBold(`"b"`), cyanBold("100"),
		blueBold(`"c"`), blackBold("null"),
		blueBold(`"d"`), yelloBold("true"),
		blueBold(`"e"`), yelloBold("false"),
		blueBold(`"f"`), blueBold(`"key"`), greenBold(`"str"`),
		blueBold(`"g"`),
		blueBold(`"h"`),
	)

	test(expected, actual)
	test("{}", prettyJson("{}"))
	test("[]", prettyJson("[]"))
}

func TestStringEscape(t *testing.T) {
	f := prettyjson.NewFormatter()
	f.DisabledColor = true
	s := `{"foo":"foo\"\nbar"}`
	r, err := f.Format([]byte(s))

	if err != nil {
		t.Error(err)
	}

	expected := `{
  "foo": "foo\"\nbar"
}`

	if string(r) != expected {
		t.Errorf("actual: %s\nexpected: %s", string(r), expected)
	}
}

func TestStringPercentEscape(t *testing.T) {
	f := prettyjson.NewFormatter()
	s := `{"foo":"foo%2Fbar"}`
	r, err := f.Format([]byte(s))

	if err != nil {
		t.Error(err)
	}

	expectedFormat := `{
  %s: %s
}`

	blueBold := color.New(color.FgBlue, color.Bold).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()

	expected := fmt.Sprintf(expectedFormat,
		blueBold(`"foo"`), greenBold(`"foo%2Fbar"`),
	)

	if string(r) != expected {
		t.Errorf("actual: %s\nexpected: %s", string(r), expected)
	}
}

func TestStringPercentEscape_DisabledColor(t *testing.T) {
	f := prettyjson.NewFormatter()
	f.DisabledColor = true
	s := `{"foo":"foo%2Fbar"}`
	r, err := f.Format([]byte(s))

	if err != nil {
		t.Error(err)
	}

	expected := `{
  "foo": "foo%2Fbar"
}`

	if string(r) != expected {
		t.Errorf("actual: %s\nexpected: %s", string(r), expected)
	}
}
