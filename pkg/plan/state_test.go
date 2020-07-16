package plan

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func p(s string) State {
	r := strings.NewReader(s)
	params, err := NewStateFromJSON(r)
	if err != nil {
		log.Fatal(s, err)
	}
	return params
}

func TestGet(t *testing.T) {
	tests := []struct {
		o        State
		path     string
		valid    bool
		expected interface{}
	}{
		{map[string]interface{}{}, "foo.bar", false, nil},
		{p(`{ "foo": 2 }`), "foo.bar", false, nil},
		{p(`{ "foo": { "bar": 2 } }`), "foo.bar", true, float64(2)},
		{p(`{ "foo": { "bar": "baz" } }`), "foo.bar", true, "baz"},
		{p(`{ "foo": { "bar": { "baz": 3 } } }`), "foo.bar", true, p(`{ "baz": 3 }`)},
		{p(`{ "xxx": "yyy", "foo": { "bar": { "baz": 3 } } }`), "foo.bar", true, p(`{ "baz": 3 }`)},
	}

	for _, test := range tests {
		v, err := test.o.Get(test.path)
		if test.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		assert.Equal(t, test.expected, v)
	}
}

func TestTypedGet(t *testing.T) {
	params := p(`{ "xxx": "yyy", "foo": { "bar": { "baz": 3 }, "boolean": true } }`)

	vBool, err := params.GetBool("foo.boolean")
	assert.NoError(t, err)
	assert.Equal(t, true, vBool)

	vNumber, err := params.GetNumber("foo.bar.baz")
	assert.NoError(t, err)
	assert.Equal(t, float64(3), vNumber)

	vString, err := params.GetString("xxx")
	assert.NoError(t, err)
	assert.Equal(t, "yyy", vString)
	_, err = params.GetString("foo.bar.baz")
	assert.Error(t, err)

	vObject, err := params.GetObject("foo.bar")
	assert.NoError(t, err)
	assert.Equal(t, p(`{ "baz": 3 }`), vObject)

}

func TestCoercion(t *testing.T) {
	params := p(`{ "n": "0.2", "b": "true" }`)

	// Happy cases.
	vNumber, err := params.GetNumber("n")
	assert.NoError(t, err)
	assert.Equal(t, 0.2, vNumber)

	vBool, err := params.GetBool("b")
	assert.NoError(t, err)
	assert.Equal(t, true, vBool)

	// Invalid coercion.
	_, err = params.GetNumber("b")
	assert.Error(t, err)
	_, err = params.GetBool("n")
	assert.Error(t, err)
}

func TestSet(t *testing.T) {
	tests := []struct {
		o        State
		path     string
		value    interface{}
		expected State
	}{
		{p(`{}`), "foo", float64(2), p(`{ "foo": 2 }`)},
		{p(`{}`), "foo", "bar", p(`{ "foo": "bar" }`)},
		{p(`{}`), "foo", true, p(`{ "foo": true }`)},
		{p(`{}`), "foo", p(`{ "bar": "baz" } `), p(`{ "foo": { "bar": "baz" } }`)},
		{p(`{ "foo": { "xxx": 42 } }`), "foo.yyy", p(`{ "bar": "baz" } `), p(`{ "foo": { "xxx": 42, "yyy": { "bar": "baz" } } }`)},
	}

	for _, test := range tests {
		test.o.Set(test.path, test.value)
		assert.Equal(t, test.expected, test.o)
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		a, b     State
		expected State
	}{
		{NewState(), NewState(), NewState()},
		{map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}},
		{map[string]interface{}{}, map[string]interface{}{"foo": 1}, map[string]interface{}{"foo": 1}},
		{map[string]interface{}{}, map[string]interface{}{"foo": "bar"}, map[string]interface{}{"foo": "bar"}},
		{map[string]interface{}{"foo": 1}, map[string]interface{}{}, map[string]interface{}{"foo": 1}},
		{map[string]interface{}{"foo": "bar"}, map[string]interface{}{}, map[string]interface{}{"foo": "bar"}},

		{p(`{ "foo": 1 } `), p(`{ "foo": { "bar": "baz" } }`), p(`{"foo": { "bar": "baz" } }`)},
		{p(`{ "foo": 1, "orig": "xxx" } `), p(`{ "foo": { "bar": "baz" } }`), p(`{"foo": { "bar": "baz" }, "orig": "xxx" }`)},
		{p(`{ "foo": { "rab": "zab" }, "orig": "xxx" } `), p(`{ "foo": { "bar": "baz" } }`), p(`{"foo": { "bar": "baz", "rab": "zab" }, "orig": "xxx" }`)},
		{p(`{ "foo": { "bar": "baz" } }`), p(`{ "foo": { "rab": "zab" }, "orig": "xxx" } `), p(`{"foo": { "bar": "baz", "rab": "zab" }, "orig": "xxx" }`)},
	}

	for _, test := range tests {
		test.a.Merge(test.b)
		assert.Equal(t, test.expected, test.a)
	}

}

func TestEqual(t *testing.T) {
	tests := []struct {
		json1    string
		json2    string
		expected bool
	}{
		{`{}`, `{}`, true},
		{`{ }`, `{ "foo": 2 }`, false},
		// Data types
		{`{ "foo": 2 }`, `{  "foo": "2" }`, false},
		{`{ "foo": "2" }`, `{  "foo": "2" }`, true},
		// Check whitespace differences don't matter
		{`{ "foo": 2 }`, `{  "foo": 2 }`, true},
		{`{ "foo": 2 }`, `{ 
			"foo": 2 }`, true},
		// Check ordering doesn't matter
		{`{ "foo": 1, "bar": 2 }`, `{"bar": 2, "foo": 1 }`, true},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			state1 := p(test.json1)
			state2 := p(test.json2)
			assert.Equal(t, test.expected, state1.Equal(state2), fmt.Sprintf("%q == %q", test.json1, test.json2))
		})
	}
}
