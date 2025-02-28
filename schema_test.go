package jsonschema

import (
	"bytes"
	"github.com/json-iterator/go"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	// "net/http"
	// "net/http/httptest"
	"path/filepath"
	"testing"
)

func ExampleBasic() {
	var schemaData = []byte(`{
    "title": "Person",
    "type": "object",
    "$comment" : "sample comment",
    "properties": {
        "firstName": {
            "type": "string"
        },
        "lastName": {
            "type": "string"
        },
        "age": {
            "description": "Age in years",
            "type": "integer",
            "minimum": 0
        },
        "friends": {
        	"type" : "array",
        	"items" : { "title" : "REFERENCE", "$ref" : "#" }
        }
    },
    "required": ["firstName", "lastName"]
	}`)

	rs := &RootSchema{}
	if err := jsoniter.Unmarshal(schemaData, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}

	var valid = []byte(`{
		"firstName" : "George",
		"lastName" : "Michael"
		}`)
	errs, err := rs.ValidateBytes(valid)
	if err != nil {
		panic(err)
	}

	var invalidPerson = []byte(`{
		"firstName" : "Prince"
		}`)

	errs, err = rs.ValidateBytes(invalidPerson)
	if err != nil {
		panic(err)
	}

	fmt.Println(errs[0].Error())

	var invalidFriend = []byte(`{
		"firstName" : "Jay",
		"lastName" : "Z",
		"friends" : [{
			"firstName" : "Nas"
			}]
		}`)
	errs, err = rs.ValidateBytes(invalidFriend)
	if err != nil {
		panic(err)
	}

	fmt.Println(errs[0].Error())

	// Output: /: {"firstName":"Prince... "lastName" value is required
	// /friends/0: {"firstName":"Nas"} "lastName" value is required
}

func TestTopLevelType(t *testing.T) {
	schemaObject := []byte(`{
    "title": "Car",
    "type": "object",
    "properties": {
        "color": {
            "type": "string"
        }
    },
    "required": ["color"]
}`)
	rs := &RootSchema{}
	if err := jsoniter.Unmarshal(schemaObject, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
	if rs.TopLevelType() != "object" {
		t.Errorf("error: schemaObject should be an object")
	}

	schemaArray := []byte(`{
    "title": "Cities",
    "type": "array",
    "items" : { "title" : "REFERENCE", "$ref" : "#" }
}`)
	rs = &RootSchema{}
	if err := jsoniter.Unmarshal(schemaArray, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
	if rs.TopLevelType() != "array" {
		t.Errorf("error: schemaArray should be an array")
	}

	schemaUnknown := []byte(`{
    "title": "Typeless",
    "items" : { "title" : "REFERENCE", "$ref" : "#" }
}`)
	rs = &RootSchema{}
	if err := jsoniter.Unmarshal(schemaUnknown, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
	if rs.TopLevelType() != "unknown" {
		t.Errorf("error: schemaUnknown should have unknown type")
	}
}

func TestParseUrl(t *testing.T) {
	// Easy case, id is a standard URL
	schemaObject := []byte(`{
    "title": "Car",
    "type": "object",
    "$id": "http://example.com/root.json"
}`)
	rs := &RootSchema{}
	if err := jsoniter.Unmarshal(schemaObject, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}

	// Tricky case, id is only a URL fragment
	schemaObject = []byte(`{
    "title": "Car",
    "type": "object",
    "$id": "#/properites/firstName"
}`)
	rs = &RootSchema{}
	if err := jsoniter.Unmarshal(schemaObject, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}

	// Another tricky case, id is only an empty fragment
	schemaObject = []byte(`{
    "title": "Car",
    "type": "object",
    "$id": "#"
}`)
	rs = &RootSchema{}
	if err := jsoniter.Unmarshal(schemaObject, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}

func TestMust(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				if err.Error() != "unexpected end of JSON input" {
					t.Errorf("expected panic error to equal: %s", "unexpected end of JSON input")
				}
			} else {
				t.Errorf("must paniced with a non-error")
			}
		} else {
			t.Errorf("expected invalid call to Must to panic")
		}
	}()

	// Valid call to Must shouldn't panic
	rs := Must(`{}`)
	if rs == nil {
		t.Errorf("expected parse of empty schema to return *RootSchema, got nil")
		return
	}

	// This should panic, checked in defer above
	Must(``)
}

func TestDraft3(t *testing.T) {
	runJSONTests(t, []string{
		"testdata/draft3/additionalItems.json",
		// TODO - not implemented:
		// "testdata/draft3/disallow.json",
		"testdata/draft3/items.json",
		"testdata/draft3/minItems.json",
		"testdata/draft3/pattern.json",
		// "testdata/draft3/refRemote.json",
		"testdata/draft3/additionalProperties.json",
		// TODO - not implemented:
		// "testdata/draft3/divisibleBy.json",
		"testdata/draft3/maxItems.json",
		"testdata/draft3/minLength.json",
		"testdata/draft3/patternProperties.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/required.json",
		"testdata/draft3/default.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/enum.json",
		"testdata/draft3/maxLength.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/minimum.json",
		"testdata/draft3/properties.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/type.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/dependencies.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/extends.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/maximum.json",
		// TODO - currently doesn't parse:
		// "testdata/draft3/ref.json",
		"testdata/draft3/uniqueItems.json",
		// "testdata/draft3/optional/bignum.json",
		// "testdata/draft3/optional/format.json",
		// "testdata/draft3/optional/jsregex.json",
		// "testdata/draft3/optional/zeroTerminatedFloats.json",
	})
}

func TestDraft4(t *testing.T) {
	runJSONTests(t, []string{
		"testdata/draft4/additionalItems.json",
		// TODO - currently doesn't parse:
		// "testdata/draft4/definitions.json",
		"testdata/draft4/maxLength.json",
		"testdata/draft4/minProperties.json",
		// "testdata/draft4/refRemote.json",
		"testdata/draft4/additionalProperties.json",
		"testdata/draft4/dependencies.json",
		"testdata/draft4/maxProperties.json",
		// TODO - currently doesn't parse:
		// "testdata/draft4/minimum.json",
		"testdata/draft4/pattern.json",
		"testdata/draft4/required.json",
		"testdata/draft4/allOf.json",
		"testdata/draft4/enum.json",
		// TODO - currently doesn't parse:
		// "testdata/draft4/maximum.json",
		"testdata/draft4/multipleOf.json",
		"testdata/draft4/patternProperties.json",
		"testdata/draft4/type.json",
		"testdata/draft4/anyOf.json",
		"testdata/draft4/items.json",
		"testdata/draft4/minItems.json",
		"testdata/draft4/not.json",
		"testdata/draft4/properties.json",
		"testdata/draft4/uniqueItems.json",
		"testdata/draft4/default.json",
		"testdata/draft4/maxItems.json",
		"testdata/draft4/minLength.json",
		"testdata/draft4/oneOf.json",
		// TODO - currently doesn't parse:
		// "testdata/draft4/ref.json",

		// "testdata/draft4/optional/bignum.json",
		// "testdata/draft4/optional/ecmascript-regex.json",
		// "testdata/draft4/optional/format.json",
		// "testdata/draft4/optional/zeroTerminatedFloats.json",
	})
}

func TestDraft6(t *testing.T) {
	runJSONTests(t, []string{
		"testdata/draft6/additionalItems.json",
		"testdata/draft6/const.json",
		"testdata/draft6/enum.json",
		"testdata/draft6/maxLength.json",
		"testdata/draft6/minProperties.json",
		"testdata/draft6/ref.json",
		"testdata/draft6/additionalProperties.json",
		"testdata/draft6/contains.json",
		"testdata/draft6/exclusiveMaximum.json",
		"testdata/draft6/maxProperties.json",
		"testdata/draft6/minimum.json",
		"testdata/draft6/pattern.json",
		// "testdata/draft6/refRemote.json",
		"testdata/draft6/allOf.json",
		"testdata/draft6/default.json",
		"testdata/draft6/exclusiveMinimum.json",
		"testdata/draft6/maximum.json",
		"testdata/draft6/multipleOf.json",
		"testdata/draft6/patternProperties.json",
		"testdata/draft6/required.json",
		"testdata/draft6/anyOf.json",
		"testdata/draft6/definitions.json",
		"testdata/draft6/items.json",
		"testdata/draft6/minItems.json",
		"testdata/draft6/not.json",
		"testdata/draft6/properties.json",
		"testdata/draft6/type.json",
		"testdata/draft6/boolean_schema.json",
		"testdata/draft6/dependencies.json",
		"testdata/draft6/maxItems.json",
		"testdata/draft6/minLength.json",
		"testdata/draft6/oneOf.json",
		"testdata/draft6/propertyNames.json",
		"testdata/draft6/uniqueItems.json",

		// "testdata/draft6/optional/bignum.json",
		// "testdata/draft6/optional/ecmascript-regex.json",
		// "testdata/draft6/optional/format.json",
		// "testdata/draft6/optional/zeroTerminatedFloats.json",
	})
}

func TestDraft7(t *testing.T) {
	prev := DefaultSchemaPool
	defer func() { DefaultSchemaPool = prev }()

	path := "testdata/draft-07_schema.json"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("error reading %s: %s", path, err.Error())
		return
	}

	rsch := &RootSchema{}
	if err := jsoniter.Unmarshal(data, rsch); err != nil {
		t.Errorf("error unmarshaling schema: %s", err.Error())
		return
	}

	DefaultSchemaPool["http://json-schema.org/draft-07/schema#"] = &rsch.Schema

	runJSONTests(t, []string{
		"testdata/draft7/additionalItems.json",
		"testdata/draft7/contains.json",
		"testdata/draft7/exclusiveMinimum.json",
		"testdata/draft7/maximum.json",
		"testdata/draft7/not.json",
		"testdata/draft7/propertyNames.json",
		"testdata/draft7/additionalProperties.json",
		"testdata/draft7/default.json",
		"testdata/draft7/if-then-else.json",
		"testdata/draft7/minItems.json",
		"testdata/draft7/oneOf.json",
		"testdata/draft7/ref.json",
		"testdata/draft7/allOf.json",
		"testdata/draft7/definitions.json",
		"testdata/draft7/items.json",
		"testdata/draft7/minLength.json",
		// "testdata/draft7/refRemote.json",
		"testdata/draft7/anyOf.json",
		"testdata/draft7/dependencies.json",
		"testdata/draft7/maxItems.json",
		"testdata/draft7/minProperties.json",
		"testdata/draft7/pattern.json",
		"testdata/draft7/required.json",
		"testdata/draft7/boolean_schema.json",
		"testdata/draft7/enum.json",
		"testdata/draft7/maxLength.json",
		"testdata/draft7/minimum.json",
		"testdata/draft7/patternProperties.json",
		"testdata/draft7/type.json",
		"testdata/draft7/const.json",
		"testdata/draft7/exclusiveMaximum.json",
		"testdata/draft7/maxProperties.json",
		"testdata/draft7/multipleOf.json",
		"testdata/draft7/properties.json",
		"testdata/draft7/uniqueItems.json",

		// "testdata/draft7/optional/bignum.json",
		// "testdata/draft7/optional/content.json",
		// "testdata/draft7/optional/ecmascript-regex.json",
		// "testdata/draft7/optional/zeroTerminatedFloats.json",
		"testdata/draft7/optional/format/date-time.json",
		"testdata/draft7/optional/format/hostname.json",
		"testdata/draft7/optional/format/ipv4.json",
		"testdata/draft7/optional/format/iri.json",
		"testdata/draft7/optional/format/relative-json-pointer.json",
		"testdata/draft7/optional/format/uri-template.json",
		"testdata/draft7/optional/format/date.json",
		"testdata/draft7/optional/format/idn-email.json",
		"testdata/draft7/optional/format/ipv6.json",
		"testdata/draft7/optional/format/json-pointer.json",
		"testdata/draft7/optional/format/time.json",
		"testdata/draft7/optional/format/uri.json",
		"testdata/draft7/optional/format/email.json",
		"testdata/draft7/optional/format/idn-hostname.json",
		"testdata/draft7/optional/format/iri-reference.json",
		"testdata/draft7/optional/format/regex.json",
		"testdata/draft7/optional/format/uri-reference.json",
	})
}

// TestSet is a json-based set of tests
// JSON-Schema comes with a lovely JSON-based test suite:
// https://github.com/json-schema-org/JSON-Schema-Test-Suite
type TestSet struct {
	Description string      `json:"description"`
	Schema      *RootSchema `json:"schema"`
	Tests       []TestCase  `json:"tests"`
}

type TestCase struct {
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Valid       bool        `json:"valid"`
}

func runJSONTests(t *testing.T, testFilepaths []string) {
	tests := 0
	passed := 0
	for _, path := range testFilepaths {
		base := filepath.Base(path)
		testSets := []*TestSet{}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("error loading test file: %s", err.Error())
			return
		}

		if err := jsoniter.Unmarshal(data, &testSets); err != nil {
			t.Errorf("error unmarshaling test set %s from JSON: %s", base, err.Error())
			return
		}

		for _, ts := range testSets {
			sc := ts.Schema
			if err := sc.FetchRemoteReferences(); err != nil {
				t.Errorf("%s: %s error fetching remote references: %s", base, ts.Description, err.Error())
				continue
			}
			for i, c := range ts.Tests {
				tests++
				got := []ValError{}
				sc.Validate("/", c.Data, &got)
				valid := len(got) == 0
				if valid != c.Valid {
					t.Errorf("%s: %s test case %d: %s. error: %s", base, ts.Description, i, c.Description, got)
				} else {
					passed++
				}
			}
		}
	}
	t.Logf("%d/%d tests passed", passed, tests)
}

func TestDataType(t *testing.T) {
	cases := []struct {
		data   interface{}
		expect string
	}{
		{nil, "null"},
		{struct{}{}, "unknown"},
		{float64(4), "integer"},
		{float64(4.5), "number"},
		{"foo", "string"},
		{map[string]interface{}{}, "object"},
		{[]interface{}{}, "array"},
	}

	for i, c := range cases {
		got := DataType(c.data)
		if got != c.expect {
			t.Errorf("case %d result mismatch. expected: '%s', got: '%s'", i, c.expect, got)
		}
	}
}

func TestJSONCoding(t *testing.T) {
	cases := []string{
		"testdata/coding/false.json",
		"testdata/coding/true.json",
		"testdata/coding/std.json",
		"testdata/coding/booleans.json",
		"testdata/coding/conditionals.json",
		"testdata/coding/numeric.json",
		"testdata/coding/objects.json",
		"testdata/coding/strings.json",
	}

	for i, c := range cases {
		data, err := ioutil.ReadFile(c)
		if err != nil {
			t.Errorf("case %d error reading file: %s", i, err.Error())
			continue
		}

		rs := &RootSchema{}
		if err := jsoniter.Unmarshal(data, rs); err != nil {
			t.Errorf("case %d error unmarshaling from json: %s", i, err.Error())
			continue
		}

		output, err := jsoniter.MarshalIndent(rs, "", "  ")
		if err != nil {
			t.Errorf("case %d error marshaling to JSON: %s", i, err.Error())
			continue
		}

		if !bytes.Equal(data, output) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(data), string(output), true)
			if len(diffs) == 0 {
				t.Logf("case %d bytes were unequal but computed no difference between results", i)
				continue
			}

			t.Errorf("case %d %s mismatch:\n", i, c)
			t.Errorf("diff:\n%s", dmp.DiffPrettyText(diffs))
			t.Errorf("expected:\n%s", string(data))
			t.Errorf("got:\n%s", string(output))
			continue
		}
	}
}

func TestValidateBytes(t *testing.T) {
	cases := []struct {
		schema string
		input  string
		errors []string
	}{
		{`true`, `"just a string yo"`, nil},
		{`{"type":"array", "items": {"type":"string"}}`,
			`[1,false,null]`,
			[]string{
				`/0: 1 type should be string`,
				`/1: false type should be string`,
				`/2: type should be string`,
			}},
	}

	for i, c := range cases {
		rs := &RootSchema{}
		if err := rs.UnmarshalJSON([]byte(c.schema)); err != nil {
			t.Errorf("case %d error parsing %s", i, err.Error())
			continue
		}

		errors, err := rs.ValidateBytes([]byte(c.input))
		if err != nil {
			t.Errorf("case %d error validating: %s", i, err.Error())
			continue
		}

		if len(errors) != len(c.errors) {
			t.Errorf("case %d: error length mismatch. expected: %d, got: %d", i, len(c.errors), len(errors))
			t.Errorf("%v", errors)
			continue
		}

		for j, e := range errors {
			if e.Error() != c.errors[j] {
				t.Errorf("case %d: validation error %d mismatch. expected: %s, got: %s", i, j, c.errors[j], e.Error())
				continue
			}
		}
	}
}

// TODO - finish remoteRef.json tests by setting up a httptest server on localhost:1234
// that uses an http.Dir to serve up testdata/remotes directory
// func testServer() {
// 	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 	}))
// }
