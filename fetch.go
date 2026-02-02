package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Rows [][]string

// FetchCSV returns the contents of source
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

	if contents, e = webFetch(source); e != nil {
		return nil, e
	}

	return str2rows(contents), nil
}

func FetchXLSX(source string) (Rows, error) {
	// fetch from web?
	if strings.Contains(source, "http") {
		var (
			s string
			e error
		)
		if s, e = webFetch(source); e != nil {
			return nil, e
		}

		tmpFile := fmt.Sprintf("%s/temp.xlsx", os.TempDir())
		if e1 := save(s, tmpFile); e1 != nil {
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

		// consider a smart split
		f = append(f, strings.Split(line, ","))
	}

	return f
}

func smartSplit(s, delim string) []string {
	inQuote := false
	var out []string
	var item []byte
	for _, b := range s {

		if b == '"' {
			inQuote = !inQuote
		}

		if !inQuote && b == ',' {
			out = append(out, string(item))
			item = nil
			continue
		}

		item = append(item, byte(b))
	}

	out = append(out, string(item))

	return out
}

func webFetch(url string) (string, error) {
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

func save(data, localFile string) error {
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
