package ragollama

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"slices"
	"time"
)

const (
	QUERY_CHUNKS = `
        SELECT chunks.path, chunks.content, embeddings.embedding
		FROM chunks
		INNER JOIN embeddings
		ON chunks.id = embeddings.id`
	MAX_SCORES = 3
)

func (ol *RagollamaClient) AnswerQuestion(question1 string) {

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", ol.dbPath)
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

Question: %v`, contextInfo, question1)

	fmt.Println("============== QUERY =====================\n", query, "\n===================================\n")
	start := time.Now()
	resp, err := ol.CreateChatCompletion(
		openai.ChatCompletionRequest{
			Model: GetOllamaModel(), // openai.GPT3Dot5Turbo,
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
