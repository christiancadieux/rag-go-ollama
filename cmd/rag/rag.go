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

	rootDir := flag.String("rootdir", "./docs", ".md docs directory")
	doClear := flag.Bool("clear", false, "clear DB table before inserting")

	dbPath := flag.String("db", "rag.db", "DB name")
	question1 := flag.String("q", defaultQuestion, "question")
	doCalculate := flag.Bool("calculate", false, "calculate embeddings and update DB")
	doAnswer := flag.Bool("answer", false, "answer question")
	doChunk := flag.Bool("chunk", false, "do chunks")
	flag.Parse()
	fmt.Println("Using LLM:", ragollama.GetOllamaUrl())

	ol := ragollama.NewRagollama(*dbPath)

	if *doChunk {
		ol.Chunks(*doClear, *rootDir)

	} else if *doAnswer {
		ol.AnswerQuestion(*question1)

	} else if *doCalculate {
		ol.CalculateEmbeddings()
	}
}
