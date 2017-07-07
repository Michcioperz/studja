package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/flosch/pongo2"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

const TARGETS_DIR = "targets"

type Block struct {
	Description, Query string
}

type Frontier struct {
	Year  int
	Value float64
}

type Target struct {
	School, Department, Class string
	KnownFrontiers            []Frontier
	Blocks                    []Block
}

type Result struct {
	Subject    string
	Extended   bool
	Percentage float64
}

type Score struct {
	SourceResult Result
	Calculation  float64
}

type CalculatedBlock struct {
	RelatedBlock Block
	SortedScores []Score
	BestValue    float64
}

type CalculatedTarget struct {
	RelatedTarget    Target
	CalculatedBlocks []CalculatedBlock
	Sum              float64
}

type Subject struct {
	Codename, FullName string
	Extended           bool
}

var Targets = []Target{}

var Subjects = []Subject{
	{"pol", "język polski", false},
	{"mat", "matematyka", false},
	{"ang", "język angielski", false},
	{"fra", "język francuski", false},
	{"hiszp", "język hiszpański", false},
	{"niem", "język niemiecki", false},
	{"ros", "język rosyjski", false},
	{"wlo", "język włoski", false},
	{"bia", "język białoruski", false},
	{"lit", "język litewski", false},
	{"ukr", "język ukraiński", false},
	{"biol", "biologia", true},
	{"chem", "chemia", true},
	{"fil", "filozofia", true},
	{"fiz", "fizyka", true},
	{"geo", "geografia", true},
	{"his", "historia", true},
	{"muz", "historia muzyki", true},
	{"szt", "historia sztuki", true},
	{"inf", "informatyka", true},
	{"ang", "język angielski", true},
	{"bia", "język białoruski", true},
	{"fra", "język francuski", true},
	{"hiszp", "język hiszpański", true},
	{"kasz", "język kaszubski", true},
	{"lit", "język litewski", true},
	{"lat", "język łaciński i kultura antyczna", true},
	{"lemk", "język łemkowski", true},
	{"niem", "język niemiecki", true},
	{"pol", "język polski", true},
	{"ros", "język rosyjski", true},
	{"ukr", "język ukraiński", true},
	{"wlo", "język włoski", true},
	{"mat", "matematyka", true},
	{"wos", "wiedza o społeczeństwie", true},
}

var FormTemplate = pongo2.Must(pongo2.FromFile("form.tpl2.html"))
var ResultsTemplate = pongo2.Must(pongo2.FromFile("results.tpl2.html"))

func main() {
	var err error
	targets, err := ioutil.ReadDir(TARGETS_DIR)
	if err != nil {
		log.Panic(err)
		return
	}
	for _, targetFileInfo := range targets {
		var target = new(Target)
		targetFileName := path.Join(TARGETS_DIR, targetFileInfo.Name())
		log.Println("loading", targetFileName)
		if path.Ext(targetFileName) == ".json" {
			targetFile, err := os.Open(targetFileName)
			if err != nil {
				log.Println("skipping file for error", err)
				continue
			}
			defer targetFile.Close()
			dec := json.NewDecoder(targetFile)
			err = dec.Decode(target)
			if err != nil {
				log.Println("skipping file for error", err)
				continue
			}
			Targets = append(Targets, *target)
		}
	}
	if len(Targets) < 1 {
		log.Panic("no targets specified, what's the point")
		return
	}
	http.HandleFunc("/oh-boi/", formReactor)
	http.HandleFunc("/results.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "results.css")
	})
	http.HandleFunc("/", formCreator)
	http.ListenAndServe(":9008", nil)
}

func formReactor(writer http.ResponseWriter, request *http.Request) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Print(err)
		fmt.Fprint(writer, "Nie udało się ogarnąć bazy danych")
		return
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE results (
  subject    TEXT    NOT NULL,
  extended   BOOLEAN NOT NULL,
  percentage FLOAT   NOT NULL
)`)
	if err != nil {
		log.Print(err)
		fmt.Fprint(writer, "Nie udało się ogarnąć bazy danych")
		return
	}
	resultInsertor, err := db.Prepare(`INSERT INTO results (subject, extended, percentage) VALUES (?, ?, ?)`)
	if err != nil {
		log.Print(err)
		fmt.Fprint(writer, "Nie udało się ogarnąć bazy danych")
		return
	}
	for _, subject := range Subjects {
		var subjectStr string
		if subject.Extended {
			subjectStr += "rozsz-"
		} else {
			subjectStr += "podst-"
		}
		subjectStr += subject.Codename
		if scoreStr := request.FormValue(subjectStr); len(scoreStr) > 0 {
			score, err := strconv.ParseFloat(scoreStr, 64)
			if err != nil {
				log.Print(err)
				fmt.Fprint(writer, "Liczby się nie dodają")
				return
			}
			_, err = resultInsertor.Exec(subject.FullName, subject.Extended, score)
			if err != nil {
				log.Print(err)
				fmt.Fprint(writer, "Nie udało się wprowadzić wyników")
				return
			}
		}
	}

	var calcTs []CalculatedTarget
	for _, target := range Targets {
		var calcT CalculatedTarget
		calcT.RelatedTarget = target
		for _, block := range target.Blocks {
			var calcB CalculatedBlock
			calcB.RelatedBlock = block
			query, err := db.Query(block.Query)
			if err != nil {
				log.Print(err)
				fmt.Fprint(writer, "Obliczenia się wywróciły")
				return
			}
			for query.Next() {
				var calcS Score
				err := query.Scan(&calcS.SourceResult.Subject, &calcS.SourceResult.Extended, &calcS.SourceResult.Percentage, &calcS.Calculation)
				if err != nil {
					log.Print(err)
					fmt.Fprint(writer, "Podobno się policzyło, ale nie dało się tego odczytać")
				}
				calcB.SortedScores = append(calcB.SortedScores, calcS)
			}
			if len(calcB.SortedScores) > 0 {
				calcB.BestValue = calcB.SortedScores[0].Calculation
			} else {
				calcB.BestValue = 0
			}
			calcT.CalculatedBlocks = append(calcT.CalculatedBlocks, calcB)
			calcT.Sum += calcB.BestValue
		}
		calcTs = append(calcTs, calcT)
	}

	err = ResultsTemplate.ExecuteWriter(pongo2.Context{"targets": calcTs}, writer)
	if err != nil {
		log.Println(err)
		http.Error(writer, "coś nie wyszło. ups.", http.StatusInternalServerError)
	}
}

func formCreator(writer http.ResponseWriter, request *http.Request) {
	err := FormTemplate.ExecuteWriter(pongo2.Context{"subjects": Subjects}, writer)
	if err != nil {
		log.Println(err)
		http.Error(writer, "coś nie wyszło. ups.", http.StatusInternalServerError)
	}
}
