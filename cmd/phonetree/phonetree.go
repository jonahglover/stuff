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

type Scout struct {
	Name       string
	Home       string
	Cell       string
	Leadership string
	Patrol     string
}

type Page struct {
	Positions []string
	Scouts    []*Scout
}

func fixPhone(s string) string {
	r := []rune(s)
	j := 0
	for i := range r {
		if r[i] == '-' || r[i] == ' ' {
			continue
		} else if '0' <= r[i] && r[i] <= '9' {
			r[j] = r[i]
			j += 1
		} else {
			return ""
		}
	}
	if j != 10 {
		return ""
	}
	return string(r[:3]) + "-" + string(r[3:6]) + "-" + string(r[6:10])
}

func main() {
	records, err := csv.NewReader(os.Stdin).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	nameIndex := column("Name", records)
	homeIndex := column("Home Phone", records)
	cellIndex := column("Cell Phone", records)
	leadershipIndex := column("Leadership", records)
	patrolIndex := column("Patrol", records)

	executiveASPL := Page{Positions: []string{
		"Flags Patrol Leader",
		"Setup Patrol Leader",
		"Quartermaster",
		"Bugler",
		"Librarian",
		"Troop Guide Lead",
	}}

	programASPL := Page{Positions: []string{
		"Scout Skills Patrol Leader",
		"Advanced Program Patrol Leader",
		"Inter Patrol Activity Patrol Leader",
	}}

	serviceASPL := Page{Positions: []string{
		"HSLC Service Patrol Leader",
		"HSLC Service Patrol Leader",
		"Order of the Arrow Representative",
		"Chaplain's Aide",
	}}

	outingsASPL := Page{Positions: []string{
		"Cleanup Patrol Leader",
		"Historian",
		"Outing Leaders current month",
		"Outing Leaders next month",
		"Outing Leaders two months out",
	}}

	communicationsASPL := Page{Positions: []string{
		"Scribe",
	}}

	pages := map[string]*Page{
		"Executive ASPL":         &executiveASPL,
		"ASPL Program":           &programASPL,
		"ASPL Community Service": &serviceASPL,
		"ASPL Outings":           &outingsASPL,
		"ASPL Communications":    &communicationsASPL,
	}

	for _, record := range records[1:] {
		s := &Scout{
			Name:   record[nameIndex],
			Home:   fixPhone(record[homeIndex]),
			Cell:   fixPhone(record[cellIndex]),
			Patrol: record[patrolIndex],
		}

		positions := map[string]bool{}
		for _, s := range strings.Split(record[leadershipIndex], ",") {
			positions[strings.TrimSuffix(strings.TrimSpace(s), " (PLC)")] = true
		}

		if positions["Patrol Leader"] {
			s.Leadership = "PL - " + s.Patrol
		}
		for p := range positions {
			if p == "Patrol Leader" {
				continue
			}
			if s.Leadership != "" {
				s.Leadership += ", "
			}
			s.Leadership += p
		}

		patrolTitle := s.Patrol + " PL"
		patrol := pages[patrolTitle]
		if patrol == nil {
			patrol = &Page{}
			pages[patrolTitle] = patrol
		}
		patrol.Scouts = append(patrol.Scouts, s)

		if positions["Patrol Leader"] ||
			positions["Bugler"] ||
			positions["Quartermaster"] ||
			positions["Librarian"] ||
			positions["Troop Guide"] {
			executiveASPL.Scouts = append(executiveASPL.Scouts, s)
		}
		if positions["Patrol Leader"] {
			programASPL.Scouts = append(programASPL.Scouts, s)
		}
		if positions["Patrol Leader"] ||
			positions["Chaplain's Aide"] ||
			positions["Order of the Arrow Representative"] {
			serviceASPL.Scouts = append(serviceASPL.Scouts, s)
		}
		if positions["Patrol Leader"] ||
			positions["Historian"] {
			outingsASPL.Scouts = append(outingsASPL.Scouts, s)
		}
		if positions["Scribe"] {
			communicationsASPL.Scouts = append(communicationsASPL.Scouts, s)
		}
	}

	if err := tmpl.Execute(os.Stdout, pages); err != nil {
		log.Fatal(err)
	}
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<meta http-equiv="content-type" content="text/html;charset=UTF-8" />
<link rel="stylesheet" href="http://gary.burd.info/site.css">
<style>
    td { padding-right: 2ex; }
</style>
<style media="print">
    html { font-size: 13px }
    div.page { page-break-after: always; page-break-inside: avoid; }
</style>
<body>
{{range $title, $page := .}}
    <div class="page">
    <h4>Phone Tree - {{$title}}</h4>
    {{with $page.Positions}}
        Scouts to call:
        <ul>
        {{range .}}
            <li>{{.}}
        {{end}}
        </ul>
    {{else}}
        Things to call about:
        <ul>
            <li>Upcoming outings
            <li>Upcoming service opportunities
            <li>Upcoming day events
            <li>Check TroopKit / Troop Web Host
            <li>Remind of next troop meeting
            <li>If PLC is the next meeting, remind them they don't need to come
            <li>Ask them if they have anything that they want brought up at the PLC
            <li>Ideas for upcoming program that patrol is responsible for
            <li>Patrol responsibilities for the current and next months
        </ul>
    {{end}}

    <table>
        <tr><td>Name</td><td>Home</td><td>Cell</td><td>Position</td></tr>
    {{range $page.Scouts}}
        <tr><td>{{.Name}}</td><td>{{.Home}}</td><td>{{.Cell}}</td><td>{{.Leadership}}</td></tr>
    {{end}}
    </table>
    </div>
{{end}}
</body>
</html>
`))
