package main

import (
	"encoding/csv"
	"html/template"
	"log"
	"os"
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

type Scout struct {
	Name       string
	Patrol     string
}

func main() {
	records, err := csv.NewReader(os.Stdin).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	nameIndex := column("Name", records)
	patrolIndex := column("Patrol", records)

    patrols := make(map[string][]string)

	for _, record := range records[1:] {
        patrol := record[patrolIndex]
        name := record[nameIndex]
        patrols[patrol] = append(patrols[patrol], name)
    }

	if err := tmpl.Execute(os.Stdout, patrols); err != nil {
		log.Fatal(err)
	}
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<meta http-equiv="content-type" content="text/html;charset=UTF-8" />
<link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css">
<style>
    td { padding-right: 2ex; }
</style>
<style media="print">
    html { font-size: 16px }
    div.page { page-break-after: always; page-break-inside: avoid; }
</style>
<body>
<div class="container">
{{range $patrol, $scouts := .}}
    <div class="page">
    <h4>{{$patrol}}</h4>
    <table class="table table-bordered">
        <tr><th>Name</th><th>Sept 8</th><th>Sept 15</th><th>Sept 29</th><th>Oct 6</th></tr>
    {{range $scouts}}
        <tr><td>{{.}}</td><td><td></td><td></td><td></td></tr>
    {{end}}
    </table>
    </div>
{{end}}
</div>
</body>
</html>
`))
