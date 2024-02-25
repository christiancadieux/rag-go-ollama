package main

import (
	"flag"
	"fmt"
	"github.com/christiancadieux/rag-go-ollama/pkg/ollama"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultQuestion = "what is a RDEI team?"
)

func main() {

	dbPath := flag.String("db", "rag.db", "DB name")
	question1 := flag.String("q", defaultQuestion, "question")
	doCalculate := flag.Bool("calculate", false, "calculate embeddings and update DB")
	doAnswer := flag.Bool("answer", false, "answer question")
	flag.Parse()
	fmt.Println("Using LLM:", ollama.GetOllamaUrl())

	ol := ollama.NewOllamaClient()

	if *doAnswer {
		ol.AnswerQuestion(*dbPath, *question1)

	} else if *doCalculate {
		ol.CalculateEmbeddings(*dbPath)
	}
}
