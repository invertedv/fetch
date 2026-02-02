package fetch

import (
	"fmt"
	"testing"

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
	assert.Equal(t, exp, smartSplit(s, ","))
}
