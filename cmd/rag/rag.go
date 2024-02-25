package main

import (
	"flag"
	"fmt"
	"github.com/christiancadieux/rag-go-ollama/pkg/ragollama"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultQuestion = "what is a RDEI team?"
)

func main() {

	doChunk := flag.Bool("chunk", false, "do chunks")
	rootDir := flag.String("rootdir", "./docs", ".md docs directory")
	doClear := flag.Bool("clear", false, "clear DB table before inserting")
	dbPath := flag.String("db", "rag.db", "DB name")

	question1 := flag.String("q", defaultQuestion, "question")
	doCalculate := flag.Bool("calculate", false, "calculate embeddings and update DB")
	doAnswer := flag.Bool("answer", false, "answer question")

	flag.Parse()

	ol := ragollama.NewRagollama(*dbPath)

	fmt.Printf("Using LLL=%s, model=%s \n", ragollama.GetOllamaUrl(), ragollama.GetOllamaModel())

	var err error

	if *doChunk {
		err = ol.Chunks(*doClear, *rootDir)
		if err != nil {
			fmt.Printf("Chunks - %v \n", err)
		}

	} else if *doCalculate {
		err = ol.CalculateEmbeddings()
		if err != nil {
			fmt.Printf("CalculateEmbeddings - %v \n", err)
		}

	} else if *doAnswer {
		err = ol.AnswerQuestion(*question1)
		if err != nil {
			fmt.Printf("AnwerQuestion - %v \n", err)
		}
	}
}
