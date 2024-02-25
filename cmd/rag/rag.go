package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chewxy/math32"
	"github.com/christiancadieux/rag-go-ollama/pkg/ollama"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
	"log"
	"slices"
	"time"
)

const (
	CREATE_EMBEDDINGS = `
         CREATE TABLE IF NOT EXISTS embeddings (
           id INTEGER PRIMARY KEY,
           embedding BLOB
		 )`
	QUERY_CHUNKS = `
        SELECT chunks.path, chunks.content, embeddings.embedding
		FROM chunks
		INNER JOIN embeddings
		ON chunks.id = embeddings.id`
	MAX_SCORES      = 3
	defaultQuestion = "what is a RDEI team?"
)

func main() {

	dbPath := flag.String("db", "rag.db", "DB name")
	question1 := flag.String("q", defaultQuestion, "question")
	doCalculate := flag.Bool("calculate", false, "calculate embeddings and update DB")
	doAnswer := flag.Bool("answer", false, "answer question")
	flag.Parse()
	fmt.Println("Using LLM:", ollama.GetOllamaUrl())

	if *doAnswer {
		answerQuestion(*dbPath, *question1)

	} else if *doCalculate {
		calculateEmbeddings(*dbPath)
	}
}

// answerQuestion is a scripted interaction with the OpenAI API using RAG.
// It takes a question (constant theQuestion), finds the most relevant
// chunks of information to it and places them in the context for the question
// to get a good answer from the LLM.
func answerQuestion(dbPath string, question1 string) {

	ol := ollama.NewOllamaClient()

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	checkErr(err)
	defer db.Close()

	// SQL query to extract chunks' content along with embeddings.
	stmt, err := db.Prepare(QUERY_CHUNKS)
	checkErr(err)
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type scoreRecord struct {
		Path    string
		Score   float32
		Content string
	}
	var scores []scoreRecord

	// Iterate through the rows, scoring each chunk with cosine similarity to
	// the question's embedding.
	qEmb := ol.GetEmbedding(question1)
	for rows.Next() {
		var (
			path      string
			content   string
			embedding []byte
		)

		err = rows.Scan(&path, &content, &embedding)
		if err != nil {
			log.Fatal(err)
		}

		contentEmb := decodeEmbedding(embedding)
		score := cosineSimilarity(qEmb, contentEmb)
		scores = append(scores, scoreRecord{path, score, content})

		fmt.Printf("path: %s, score: %v, content: %d, embedding: %d\n", path, score, len(content), len(embedding))
		// fmt.Println(path, score)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	slices.SortFunc(scores, func(a, b scoreRecord) int {
		// The scores are in the range [0, 1], so scale them to get non-zero
		// integers for comparison.
		return int(100.0 * (a.Score - b.Score))
	})

	// Take the 3 best-scoring chunks as context and paste them together into
	// contextInfo.
	var contextInfo string

	if len(scores) == 0 {
		fmt.Println("No scores found")
		return
	}

	cnt := 0
	for i := len(scores) - 1; i >= 0; i-- {
		// fmt.Printf("Score %d - %v - %s\n", i, scores[i].Score, scores[i].Path)
		contextInfo = contextInfo + "\n" + scores[i].Content
		cnt++
		if cnt >= MAX_SCORES {
			break
		}
	}

	// Build the prompt and execute the LLM API.
	query := fmt.Sprintf(`Use the below information to answer the subsequent question.
Information:
%v

Question: %v`, contextInfo, defaultQuestion)

	fmt.Println("============== QUERY =====================\n", query, "\n===================================\n")
	start := time.Now()
	resp, err := ol.CreateChatCompletion(
		openai.ChatCompletionRequest{
			Model: ollama.GetOllamaModel(), // openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)
	checkErr(err)

	fmt.Println("Got response, ID:", resp.ID, "Duration:", time.Now().Sub(start))
	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	choice := resp.Choices[0]
	fmt.Println("Response.choices[0]:\n" + choice.Message.Content)
}

// calculateEmbeddings calculates embeddings for all chunks listed in the
// given DB using the OpenAI API, and stores them back into the "embeddings"
// table in the same DB.
func calculateEmbeddings(dbPath string) {

	ol := ollama.NewOllamaClient()

	db, err := sql.Open("sqlite3", dbPath)
	checkErr(err)
	defer db.Close()

	log.Println("Creating embeddings table if needed")
	_, err = db.Exec(CREATE_EMBEDDINGS)
	checkErr(err)

	log.Println("Clearing embeddings table")
	_, err = db.Exec(`DELETE FROM embeddings`)
	checkErr(err)

	rows, err := db.Query("SELECT * FROM chunks")
	checkErr(err)
	defer rows.Close()

	// Step 1: calculate embeddings for all chunks in the DB, storing them in
	// embs.
	type embData struct {
		id   int
		data []byte
	}
	var embs []embData

	for rows.Next() {

		var (
			id      int
			path    string
			nchunk  int
			content string
		)
		err = rows.Scan(&id, &path, &nchunk, &content)
		checkErr(err)

		fmt.Printf("id: %d, path: %s, nchunk: %d, content: %d\n", id, path, nchunk, len(content))
		if len(content) > 0 {
			emb := encodeEmbedding(ol.GetEmbedding(content))
			embs = append(embs, embData{id, emb})
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	rows.Close()

	// Step 2: insert all embedding data into the embeddings table.
	for _, emb := range embs {
		fmt.Println("Inserting into embeddings, id", emb.id)
		_, err = db.Exec("INSERT INTO embeddings VALUES (?, ?)", emb.id, emb.data)
		checkErr(err)
	}
}

// encodeEmbedding encodes an embedding into a byte buffer, e.g. for DB
// storage as a blob.
func encodeEmbedding(emb []float32) []byte {
	buf := new(bytes.Buffer)
	for _, f := range emb {
		err := binary.Write(buf, binary.LittleEndian, f)
		checkErr(err)
	}
	return buf.Bytes()
}

// decodeEmbedding decodes an embedding back from a byte buffer.
func decodeEmbedding(b []byte) []float32 {
	var numbers []float32
	buf := bytes.NewReader(b)

	// Calculate how many float32 values are in the slice
	count := buf.Len() / 4

	for i := 0; i < count; i++ {
		var num float32
		err := binary.Read(buf, binary.LittleEndian, &num)
		checkErr(err)
		numbers = append(numbers, num)
	}
	return numbers
}

// cosineSimilarity calculates cosine similarity (magnitude-adjusted dot
// product) between two vectors that must be of the same size.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("different lengths")
	}

	var aMag, bMag, dotProduct float32
	for i := 0; i < len(a); i++ {
		aMag += a[i] * a[i]
		bMag += b[i] * b[i]
		dotProduct += a[i] * b[i]
	}
	return dotProduct / (math32.Sqrt(aMag) * math32.Sqrt(bMag))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
