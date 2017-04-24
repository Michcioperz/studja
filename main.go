package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "html/template"
  "net/http"
  "fmt"
  "log"
  "strconv"
)

type Block struct {
  Description, Query string
}

type Target struct {
  School, Department, Class string
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

var Targets = []Target{
  {
    School:     "Uniwersytet Warszawski",
    Department: "Wydział Matematyki, Informatyki i Mechaniki",
    Class:      "Informatyka",
    Blocks: []Block{
      {
        Description: "język polski",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.6 END) * 0.1 AS score
FROM results
WHERE subject = 'język polski'
ORDER BY score DESC`,
      },
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.6 END) * 0.1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "język obcy nowożytny",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.6 END) * 0.1 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
      {
        Description: "rozszerzona matematyka lub informatyka",
        Query: `
SELECT
  *,
  percentage * 1 * 0.5 AS score
FROM results
WHERE extended = 1 AND subject IN ('matematyka', 'informatyka')
ORDER BY score DESC`,
      },
      {
        Description: "dowolne rozszerzenie",
        Query: `
SELECT
  *,
  percentage * 1 * 0.2 AS score
FROM results
WHERE extended = 1
ORDER BY score DESC`,
      },
    },
  },
}

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

func main() {
  http.HandleFunc("/oh-boi/", formReactor)
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

  templ, err := template.ParseFiles("resultcard_head.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  templ.Execute(writer, nil)

  for _, target := range calcTs {
    templ, err := template.ParseFiles("target_head.tpl.html")
    if err != nil {
      log.Print(err)
      fmt.Fprint(writer, "Nasz szablon jest zepsuty")
      return
    }
    templ.Execute(writer, target)
    for _, block := range target.CalculatedBlocks {
      templ, err := template.ParseFiles("block_head.tpl.html")
      if err != nil {
        log.Print(err)
        fmt.Fprint(writer, "Nasz szablon jest zepsuty")
        return
      }
      templ.Execute(writer, block)

      for _, score := range block.SortedScores {
        templ, err := template.ParseFiles("score.tpl.html")
        if err != nil {
          log.Print(err)
          fmt.Fprint(writer, "Nasz szablon jest zepsuty")
          return
        }
        templ.Execute(writer, score)
      }

      templ, err = template.ParseFiles("block_foot.tpl.html")
      if err != nil {
        log.Print(err)
        fmt.Fprint(writer, "Nasz szablon jest zepsuty")
        return
      }
      templ.Execute(writer, block)
    }
    templ, err = template.ParseFiles("target_foot.tpl.html")
    if err != nil {
      log.Print(err)
      fmt.Fprint(writer, "Nasz szablon jest zepsuty")
      return
    }
    templ.Execute(writer, target)
  }

  templ, err = template.ParseFiles("resultcard_foot.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  templ.Execute(writer, nil)
}

func formCreator(writer http.ResponseWriter, request *http.Request) {
  templ, err := template.ParseFiles("head.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  templ.Execute(writer, nil)
  ptempl, err := template.ParseFiles("fieldp.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  rtempl, err := template.ParseFiles("fieldr.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  for _, subject := range Subjects {
    if subject.Extended {
      rtempl.Execute(writer, subject)
    } else {
      ptempl.Execute(writer, subject)
    }
  }
  templ, err = template.ParseFiles("foot.tpl.html")
  if err != nil {
    log.Print(err)
    fmt.Fprint(writer, "Nasz szablon jest zepsuty")
    return
  }
  templ.Execute(writer, nil)
}
