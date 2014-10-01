package main

import (
	"encoding/csv"
	"html/template"
	"log"
	"os"
	"strings"
)

func column(name string, records [][]string) int {
	for i, n := range records[0] {
		if n == name {
			return i
		}
	}
	log.Fatalf("Column %s not found", name)
	return 0
}

func main() {
	records, err := csv.NewReader(os.Stdin).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	nameIndex := column("Name", records)
	records = records[1:]

	var scouts []string
	for _, record := range records {
		n := strings.Split(record[nameIndex], "\"")[0]
		if n == "" {
			continue
		}
		scouts = append(scouts, n)
	}

	split1 := len(scouts) / 3
	split2 := 2 * len(scouts) / 3
    cols := [][]string{scouts[:split1], scouts[split1:split2], scouts[split2:]}

	if err := tmpl.Execute(os.Stdout, cols); err != nil {
		log.Fatal(err)
	}
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<meta http-equiv="content-type" content="text/html;charset=UTF-8" />
<link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css">
<body>
<div class="container">
    <div class="row">
        <h3 class="text-center">Troop Meeting Attendance</h3>
    </div>
    <div class="row">
        {{range .}}
            <div class="col-xs-4">
                <table class="table table-condensed">
                    {{range .}}<tr><td>&#9633;</td><td>{{.}}</td></tr>{{end}}
                </table>
            </div>
        {{end}}
    </div>
</div>
</body>
</html>
`))
