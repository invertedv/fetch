package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchCSV(t *testing.T) {
	file := "/home/will/Downloads/DGS10.csv"
	r, e := FetchCSV(file)
	assert.Nil(t, e)
	_ = r
}

func TestFetchXLSX(t *testing.T) {
	file := "/home/will/Downloads/hpi_at_state.xlsx"
	r, e := FetchXLSX(file)
	assert.Nil(t, e)
	_ = r
}

func TestSmartSplit(t *testing.T) {
	s := `hello,"good,bye"`
	exp := []string{"hello", `"good,bye"`}
	assert.Equal(t, exp, smartSplit(s, ','))
}

func TestParseRow(t *testing.T) {
	r := []string{"2022-07-17", "43.3"}
	m := []bool{false, false}
	template := []string{"date", "float"}

	v, e := ParseRow(r, template, m, "2006-01-02")
	assert.Nil(t, e)
	exp := []any{time.Date(2022, 07, 17, 0, 0, 0, 0, time.UTC), float32(43.3)}
	act := []any{*(v[0]).(*time.Time), *v[1].(*float32)}
	assert.Equal(t, exp, act)

	template = []string{"dateCCYYMMDD", "float"}
	v, e = ParseRow(r, template, m, "2006-01-02")
	assert.Nil(t, e)
	exp = []any{20220717, float32(43.3)}
	act = []any{*(v[0]).(*int), *v[1].(*float32)}
	assert.Equal(t, exp, act)

	template = []string{"dateCCYYMMDD", "int"}
	v, e = ParseRow(r, template, m, "2006-01-02")
	assert.NotNil(t, e)

	m[1] = true
	v, e = ParseRow(r, template, m, "2006-01-02")
	assert.Nil(t, e)
	assert.Nil(t, v[1])
}
