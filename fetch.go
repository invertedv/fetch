package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// TODO: add Save?
type Rows [][]string

// FetchCSV returns the contents of source CSV
//
// - source -- either a web address or local file
func FetchCSV(source string) (Rows, error) {
	if !strings.Contains(source, "http") {
		var (
			e     error
			file  *os.File
			bytes []byte
		)

		defer file.Close()
		if file, e = os.Open(source); e != nil {
			return nil, e
		}

		if bytes, e = io.ReadAll(file); e != nil {
			return nil, e
		}

		return str2rows(string(bytes)), nil
	}

	var (
		contents string
		e        error
	)

	if contents, e = WebFetch(source); e != nil {
		return nil, e
	}

	return str2rows(contents), nil
}

// FetchXLSX returns the contents of source XLSX
//
// - source -- either a web address or local file
func FetchXLSX(source string) (Rows, error) {
	// fetch from web?
	if strings.Contains(source, "http") {
		var (
			s string
			e error
		)
		if s, e = WebFetch(source); e != nil {
			return nil, e
		}

		tmpFile := fmt.Sprintf("%s/temp.xlsx", os.TempDir())
		if e1 := Save(s, tmpFile); e1 != nil {
			return nil, e
		}
		defer os.Remove(tmpFile)

		source = tmpFile
	}

	xlr, e := excelize.OpenFile(source)
	if e != nil {
		return nil, e
	}
	defer xlr.Close()

	var (
		r  Rows
		e2 error
	)
	if r, e2 = xlr.GetRows(xlr.GetSheetName(0)); e2 != nil {
		return nil, e2
	}

	return r, nil
}

// ParseRow will parse a row of the type Rows.
//
// row      - an element of Rows
// template - expected type of each element: float, int, date, CCYYMMDD or string
// missOK   - whether the element can be missing.
// dateFormat - format for dates
//
// out - slice of pointers to parsed elements of row
func ParseRow(row, template []string, missOK []bool, dateFormat string) (out []any, e error) {
	if len(row) != len(template) || len(row) != len(missOK) {
		return nil, nil
	}

	for j, r := range row {
		var (
			x    any
			miss bool
		)

		switch template[j] {
		case "float":
			x, miss = toFloat(r)
		case "int":
			x, miss = toInt(r)
		case "date":
			x, miss = toDate(r, dateFormat)
		case "dateCCYYMMDD":
			x, miss = toCCYYMMDD(r, dateFormat)
		case "string":
			x, miss = r, false
		default:
			panic("unknown template type")
		}

		if miss && !missOK[j] {
			return nil, fmt.Errorf("missing required field")
		}

		out = append(out, x)
	}

	return out, nil
}

// Save saves the string returned by WebFetch.  Works for both CSV and XLSX.
func Save(data, localFile string) error {
	var (
		e    error
		file *os.File
	)

	if file, e = os.Create(localFile); e != nil {
		return e
	}
	defer file.Close()

	_, e = file.WriteString(data)

	return e
}

// WebFetch returns the contents of the page specified by url.
func WebFetch(url string) (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	r, _ := client.Do(req)
	defer func() { _ = r.Body.Close() }()

	var (
		body []byte
		e    error
	)
	if body, e = io.ReadAll(r.Body); e != nil {
		return "", e
	}

	return string(body), nil
}


func toFloat(s string) (*float32, bool) {
	x, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return nil, true
	}

	xx := float32(x)
	return &xx, false
}

func toInt(s string) (*int, bool) {
	x, e := strconv.ParseInt(s, 10, 64)
	if e != nil {
		return nil, true
	}

	xx := int(x)
	return &xx, false
}

func toDate(s, format string) (*time.Time, bool) {
	x, e := time.Parse(format, s)
	if e != nil {
		return nil, true
	}

	return &x, false
}

func toCCYYMMDD(s, format string) (*int, bool) {
	var d *time.Time
	if d, _ = toDate(s, format); d == nil {
		return nil, true
	}

	dd := 10000*d.Year() + 100*int(d.Month()) + d.Day()
	return &dd, false
}

func str2rows(s string) Rows {
	var f Rows

	for {
		if len(s) == 0 {
			break
		}

		indx := strings.Index(s, "\n")
		line := s
		if indx >= 0 {
			line = s[:indx]
			s = s[indx+1:]
		}

		if len(line) == 0 {
			continue
		}

		f = append(f, strings.Split(line, ","))
	}

	return f
}

// smartSplit ignores delim if it's between double quotes
func smartSplit(s string, delim rune) []string {
	inQuote := false
	var out []string
	var item []byte
	for _, b := range s {

		if b == '"' {
			inQuote = !inQuote
		}

		if !inQuote && b == delim {
			out = append(out, string(item))
			item = nil
			continue
		}

		item = append(item, byte(b))
	}

	out = append(out, string(item))

	return out
}
