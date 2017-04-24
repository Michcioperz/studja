package main

import (
  "database/sql"
  "fmt"
  _ "github.com/mattn/go-sqlite3"
  "html/template"
  "log"
  "net/http"
  "strconv"
)

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

var Targets = []Target{
  {
    School:     "Uniwersytet Warszawski",
    Department: "Wydział Matematyki, Informatyki i Mechaniki",
    Class:      "Informatyka",
    KnownFrontiers: []Frontier{
      {
        Year:  2014,
        Value: 84.18,
      },
      {
        Year:  2015,
        Value: 86.64,
      },
      {
        Year:  2016,
        Value: 85.44,
      },
    },
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
  {
    School:     "Politechnika Lubelska",
    Department: "Wydział Elektrotechniki i Informatyki",
    Class:      "Informatyka",
    Blocks: []Block{
      {
        Description: "język polski",
        Query: `
SELECT
  *,
  (CASE WHEN (extended = 1 AND percentage >= 30) THEN percentage + (6*percentage + 100)/7 ELSE percentage END) * 0.1 AS score
FROM results
WHERE subject = 'język polski'
ORDER BY score DESC`,
      },
      {
        Description: "język obcy nowożytny",
        Query: `
SELECT
  *,
  (CASE WHEN (extended = 1 AND percentage >= 30) THEN percentage + (6*percentage + 100)/7 ELSE percentage END) * 0.3 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
      {
        Description: "przedmiot magiczny",
        Query: `SELECT *
FROM (SELECT
        *,
        (CASE WHEN (extended = 1 AND percentage >= 30)
          THEN percentage + (6 * percentage + 100) / 7
         ELSE percentage END) AS score
      FROM results
      WHERE subject IN ('matematyka', 'fizyka', 'informatyka')
      UNION
      SELECT
        *,
        (CASE WHEN (extended = 1 AND percentage >= 30)
          THEN (6 * percentage + 100) / 7
         ELSE percentage END) AS score
      FROM results
      WHERE subject IN ('historia', 'biologia', 'chemia', 'geografia', 'wiedza o społeczeństwie'))
ORDER BY score
  DESC`,
      },
    },
  },
  {
    School:     "Uniwersytet Marii Curie-Skłodowskiej w Lublinie",
    Department: "Wydział Matematyki, Fizyki i Informatyki",
    Class:      "Informatyka",
    Blocks: []Block{
      {
        Description: "matematyka * 1",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 2 ELSE 1 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "fizyka * 0.6",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 2 ELSE 1 END) * 0.6 AS score
FROM results
WHERE subject = 'fizyka'
ORDER BY score DESC`,
      },
      {
        Description: "informatyka * 0.6",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 2 ELSE 1 END) * 0.6 AS score
FROM results
WHERE subject = 'informatyka'
ORDER BY score DESC`,
      },
    },
  },
  {
    School:     "Politechnika Warszawska",
    Department: "Wydział Elektroniki i Technik Informacyjnych",
    Class:      "Informatyka",
    KnownFrontiers: []Frontier{
      {
        Year:  2013,
        Value: 179,
      },
      {
        Year:  2014,
        Value: 173,
      },
      {
        Year:  2015,
        Value: 178,
      },
      {
        Year:  2016,
        Value: 184,
      },
    },
    Blocks: []Block{
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "wybrany przedmiot",
        Query: `SELECT
  *,
  percentage * (CASE WHEN extended = 1
    THEN 1
                ELSE 0.5 END) * (CASE subject
                                 WHEN 'fizyka'
                                   THEN 1
                                 WHEN 'informatyka'
                                   THEN 1
                                 WHEN 'chemia'
                                   THEN 0.75
                                 WHEN 'biologia'
                                   THEN 0.5
                                 ELSE 0 END) AS score
                                            FROM results
                                            WHERE subject IN ('fizyka', 'informatyka', 'chemia', 'biologia')
                                            ORDER BY score DESC`,
      },
      {
        Description: "język obcy * 0.25",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 0.25 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
    },
  },
  {
    School:     "Politechnika Warszawska",
    Department: "Wydział Elektryczny",
    Class:      "Informatyka",
    KnownFrontiers: []Frontier{
      {
        Year:  2013,
        Value: 153,
      },
      {
        Year:  2014,
        Value: 158,
      },
      {
        Year:  2015,
        Value: 166,
      },
      {
        Year:  2016,
        Value: 173,
      },
    },
    Blocks: []Block{
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "wybrany przedmiot",
        Query: `SELECT
  *,
  percentage * (CASE WHEN extended = 1
    THEN 1
                ELSE 0.5 END) * (CASE subject
                                 WHEN 'fizyka'
                                   THEN 1
                                 WHEN 'informatyka'
                                   THEN 1
                                 WHEN 'chemia'
                                   THEN 0.75
                                 WHEN 'biologia'
                                   THEN 0.5
                                 ELSE 0 END) AS score
                                            FROM results
                                            WHERE subject IN ('fizyka', 'informatyka', 'chemia', 'biologia')
                                            ORDER BY score DESC`,
      },
      {
        Description: "język obcy * 0.25",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 0.25 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
    },
  },
  {
    School:     "Politechnika Warszawska",
    Department: "Wydział Matematyki i Nauk Informacyjnych",
    Class:      "Informatyka",
    KnownFrontiers: []Frontier{
      {
        Year:  2013,
        Value: 182,
      },
      {
        Year:  2014,
        Value: 180,
      },
      {
        Year:  2015,
        Value: 182,
      },
      {
        Year:  2016,
        Value: 194,
      },
    },
    Blocks: []Block{
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "wybrany przedmiot",
        Query: `SELECT
  *,
  percentage * (CASE WHEN extended = 1
    THEN 1
                ELSE 0.5 END) * (CASE subject
                                 WHEN 'fizyka'
                                   THEN 1
                                 WHEN 'informatyka'
                                   THEN 1
                                 WHEN 'chemia'
                                   THEN 0.75
                                 WHEN 'biologia'
                                   THEN 0.5
                                 ELSE 0 END) AS score
                                            FROM results
                                            WHERE subject IN ('fizyka', 'informatyka', 'chemia', 'biologia')
                                            ORDER BY score DESC`,
      },
      {
        Description: "język obcy * 0.25",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 0.25 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
    },
  },
  {
    School:     "Politechnika Warszawska",
    Department: "Wydział Elektroniki i Technik Informacyjnych",
    Class:      "Computer Science",
    KnownFrontiers: []Frontier{
      {
        Year:  2015,
        Value: 124,
      },
      {
        Year:  2016,
        Value: 145,
      },
    },
    Blocks: []Block{
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "wybrany przedmiot",
        Query: `SELECT
  *,
  percentage * (CASE WHEN extended = 1
    THEN 1
                ELSE 0.5 END) * (CASE subject
                                 WHEN 'fizyka'
                                   THEN 1
                                 WHEN 'informatyka'
                                   THEN 1
                                 WHEN 'chemia'
                                   THEN 0.75
                                 WHEN 'biologia'
                                   THEN 0.5
                                 ELSE 0 END) AS score
                                            FROM results
                                            WHERE subject IN ('fizyka', 'informatyka', 'chemia', 'biologia')
                                            ORDER BY score DESC`,
      },
      {
        Description: "język obcy * 0.25",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 0.25 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
ORDER BY score DESC`,
      },
    },
  },
  {
    School:     "Politechnika Warszawska",
    Department: "Wydział Matematyki i Nauk Informacyjnych",
    Class:      "Computer Science",
    KnownFrontiers: []Frontier{
      {
        Year:  2013,
        Value: 81,
      },
      {
        Year:  2014,
        Value: 130,
      },
      {
        Year:  2015,
        Value: 126,
      },
      {
        Year:  2016,
        Value: 157,
      },
    },
    Blocks: []Block{
      {
        Description: "matematyka",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 1 AS score
FROM results
WHERE subject = 'matematyka'
ORDER BY score DESC`,
      },
      {
        Description: "wybrany przedmiot",
        Query: `SELECT
  *,
  percentage * (CASE WHEN extended = 1
    THEN 1
                ELSE 0.5 END) * (CASE subject
                                 WHEN 'fizyka'
                                   THEN 1
                                 WHEN 'informatyka'
                                   THEN 1
                                 WHEN 'chemia'
                                   THEN 0.75
                                 WHEN 'biologia'
                                   THEN 0.5
                                 ELSE 0 END) AS score
                                            FROM results
                                            WHERE subject IN ('fizyka', 'informatyka', 'chemia', 'biologia')
                                            ORDER BY score DESC`,
      },
      {
        Description: "język obcy * 0.25",
        Query: `
SELECT
  *,
  percentage * (CASE WHEN extended = 1 THEN 1 ELSE 0.5 END) * 0.25 AS score
FROM results
WHERE subject IN
      ('język angielski', 'język francuski', 'język niemiecki', 'język hiszpański', 'język włoski', 'język rosyjski')
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

var Templates = map[string]*template.Template{}

func main() {
  for _, templ_name := range []string{"block_foot.tpl.html", "block_head.tpl.html", "blocks_foot.tpl.html", "blocks_head.tpl.html", "fieldp.tpl.html", "fieldr.tpl.html", "foot.tpl.html", "frontier.tpl.html", "frontiers_foot.tpl.html", "frontiers_head.tpl.html", "head.tpl.html", "resultcard_foot.tpl.html", "resultcard_head.tpl.html", "score.tpl.html", "target_foot.tpl.html", "target_head.tpl.html"} {
    var err error
    Templates[templ_name], err = template.ParseFiles(templ_name)
    if err != nil {
      log.Panic(err)
      return
    }
  }
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

  Templates["resultcard_head.tpl.html"].Execute(writer, nil)

  for _, target := range calcTs {
    Templates["target_head.tpl.html"].Execute(writer, target)
    Templates["blocks_head.tpl.html"].Execute(writer, target)
    for _, block := range target.CalculatedBlocks {
      Templates["block_head.tpl.html"].Execute(writer, block)

      for _, score := range block.SortedScores {
        Templates["score.tpl.html"].Execute(writer, score)
      }

      Templates["block_foot.tpl.html"].Execute(writer, block)
    }
    Templates["blocks_foot.tpl.html"].Execute(writer, target)
    if len(target.RelatedTarget.KnownFrontiers) > 0 {
      Templates["frontiers_head.tpl.html"].Execute(writer, target)
      for _, frontier := range target.RelatedTarget.KnownFrontiers {
        Templates["frontier.tpl.html"].Execute(writer, frontier)
      }
      Templates["frontiers_foot.tpl.html"].Execute(writer, target)
    }
    Templates["target_foot.tpl.html"].Execute(writer, target)
  }

  Templates["resultcard_foot.tpl.html"].Execute(writer, nil)
}

func formCreator(writer http.ResponseWriter, request *http.Request) {
  Templates["head.tpl.html"].Execute(writer, nil)
  for _, subject := range Subjects {
    if subject.Extended {
      Templates["fieldp.tpl.html"].Execute(writer, subject)
    } else {
      Templates["fieldr.tpl.html"].Execute(writer, subject)
    }
  }
  Templates["foot.tpl.html"].Execute(writer, nil)
}
