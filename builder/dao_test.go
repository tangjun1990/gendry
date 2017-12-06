package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEq(t *testing.T) {
	var testData = []struct {
		in     map[string]interface{}
		outCon []string
		outVal []interface{}
	}{
		{
			map[string]interface{}{
				"foo": "bar",
				"baz": 1,
				"qq":  "ttx",
			},
			[]string{"baz=?", "foo=?", "qq=?"},
			[]interface{}{1, "bar", "ttx"},
		},
	}
	ass := assert.New(t)
	for _, testCase := range testData {
		cond, vals := Eq(testCase.in).Build()
		ass.Equal(len(cond), len(vals))
		ass.Equal(testCase.outCon, cond)
		ass.Equal(testCase.outVal, vals)
	}
}

func TestIn(t *testing.T) {
	var testData = []struct {
		in      map[string][]interface{}
		outCond []string
		outVals []interface{}
	}{
		{
			in: map[string][]interface{}{
				"foo": []interface{}{"bar", "baz"},
				"age": []interface{}{5, 7, 9, 11},
			},
			outCond: []string{"age IN (?,?,?,?)", "foo IN (?,?)"},
			outVals: []interface{}{5, 7, 9, 11, "bar", "baz"},
		},
	}
	ass := assert.New(t)
	for _, testCase := range testData {
		cond, vals := In(testCase.in).Build()
		ass.Equal(testCase.outCond, cond)
		ass.Equal(testCase.outVals, vals)
	}
}

func TestResolveFields(t *testing.T) {
	ass := assert.New(t)
	m := map[string]interface{}{
		"foo": 1,
		"bar": 2,
		"qq":  3,
		"asd": 4,
	}
	res := resolveFields(m)
	var assertion []string
	defaultSortAlgorithm(append(assertion, "foo", "bar", "qq", "asd"))
	for i := 0; i < len(assertion); i++ {
		ass.Equal(assertion[i], res[i])
	}
}

func TestAssembleExpression(t *testing.T) {
	var data = []struct {
		inField, inOp string
		out           string
	}{
		{"foo", "=", "foo=?"},
		{"qq", "<>", "qq<>?"},
	}
	ass := assert.New(t)
	for _, tc := range data {
		ass.Equal(tc.out, assembleExpression(tc.inField, tc.inOp))
	}
}

func TestOrderBy(t *testing.T) {
	var data = []struct {
		inField, inOrder string
		outStr           string
		outErr           error
	}{
		{"age", "desc", "age DESC", nil},
		{"name", "Asc", "name ASC", nil},
		{"tt", "DesC", "tt DESC", nil},
		{"qq", "DESCC", "", errOrderByParam},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actual, err := orderBy(tc.inField, tc.inOrder)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, actual)
	}
}

func TestResolveKV(t *testing.T) {
	var data = []struct {
		in      map[string]interface{}
		outStr  []string
		outVals []interface{}
	}{
		{
			map[string]interface{}{
				"foo": "bar",
				"bar": 1,
			},
			[]string{"bar", "foo"},
			[]interface{}{1, "bar"},
		},
		{
			map[string]interface{}{
				"qq":    "ttt",
				"some":  123,
				"other": 456,
			},
			[]string{"other", "qq", "some"},
			[]interface{}{456, "ttt", 123},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		keys, vals := resolveKV(tc.in)
		ass.Equal(tc.outStr, keys)
		ass.Equal(tc.outVals, vals)
	}
}

func TestWhereConnector(t *testing.T) {
	var data = []struct {
		in      []Comparable
		outStr  string
		outVals []interface{}
	}{
		{
			in: []Comparable{
				Eq(map[string]interface{}{
					"a": "a",
					"b": "b",
				}),
				Ne(map[string]interface{}{
					"foo": 1,
					"sex": "male",
				}),
				In(map[string][]interface{}{
					"qq": []interface{}{7, 8, 9},
				}),
			},
			outStr:  "(a=? AND b=? AND foo!=? AND sex!=? AND qq IN (?,?,?))",
			outVals: []interface{}{"a", "b", 1, "male", 7, 8, 9},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals := whereConnector(tc.in...)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildInsert(t *testing.T) {
	var data = []struct {
		table   string
		data    []map[string]interface{}
		outStr  string
		outVals []interface{}
		outErr  error
	}{
		{
			table: "tb1",
			data: []map[string]interface{}{
				map[string]interface{}{
					"foo": 1,
					"bar": 2,
				},
				map[string]interface{}{
					"foo": 3,
					"bar": 4,
				},
				map[string]interface{}{
					"foo": 5,
					"bar": 6,
				},
			},
			outStr:  "INSERT INTO tb1 (bar,foo) VALUES (?,?),(?,?),(?,?)",
			outVals: []interface{}{2, 1, 4, 3, 6, 5},
			outErr:  nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals, err := buildInsert(tc.table, tc.data)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildUpdate(t *testing.T) {
	var data = []struct {
		table      string
		conditions []Comparable
		data       map[string]interface{}
		outErr     error
		outStr     string
		outVals    []interface{}
	}{
		{
			table: "tb",
			conditions: []Comparable{
				Eq(map[string]interface{}{
					"foo": "bar",
					"qq":  1,
				}),
			},
			data: map[string]interface{}{
				"name": "deen",
				"age":  23,
			},
			outErr:  nil,
			outStr:  "UPDATE tb SET age=?,name=? WHERE (foo=? AND qq=?)",
			outVals: []interface{}{23, "deen", "bar", 1},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildUpdate(tc.table, tc.data, tc.conditions...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}

func TestBuildDelete(t *testing.T) {
	var data = []struct {
		table   string
		where   []Comparable
		outStr  string
		outVals []interface{}
		outErr  error
	}{
		{
			table: "tb",
			where: []Comparable{
				Eq(map[string]interface{}{
					"foo": 1,
					"bar": 2,
					"baz": "tt",
				}),
			},
			outStr:  "DELETE FROM tb WHERE (bar=? AND baz=? AND foo=?)",
			outVals: []interface{}{2, "tt", 1},
			outErr:  nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals, err := buildDelete(tc.table, tc.where...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildSelect(t *testing.T) {
	var data = []struct {
		table      string
		fields     []string
		conditions []Comparable
		groupBy    string
		orderBy    *eleOrderBy
		limit      *eleLimit
		outStr     string
		outVals    []interface{}
		outErr     error
	}{
		{
			table:  "tb",
			fields: []string{"foo", "bar"},
			conditions: []Comparable{
				Eq(map[string]interface{}{
					"foo": 1,
					"bar": 2,
				}),
				In(map[string][]interface{}{
					"qq": []interface{}{4, 5, 6},
				}),
			},
			groupBy: "",
			orderBy: &eleOrderBy{
				field: "foo",
				order: "desc",
			},
			limit: &eleLimit{
				begin: 10,
				step:  20,
			},
			outErr:  nil,
			outStr:  "SELECT foo,bar FROM tb WHERE (bar=? AND foo=? AND qq IN (?,?,?)) ORDER BY foo DESC LIMIT 10,20",
			outVals: []interface{}{2, 1, 4, 5, 6},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildSelect(tc.table, tc.fields, tc.groupBy, tc.orderBy, tc.limit, tc.conditions...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}