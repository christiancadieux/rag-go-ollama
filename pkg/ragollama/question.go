package ragollama

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
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

func (ol *RagollamaClient) getRows() (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", ol.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// SQL query to extract chunks' content along with embeddings.
	stmt, err := db.Prepare(QUERY_CHUNKS)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type scoreRecord struct {
	Path    string
	Score   float32
	Content string
}

func (ol *RagollamaClient) getContextInfo(qEmb []float32, rows *sql.Rows, max_scores int) (string, error) {

	var scores []scoreRecord

	for rows.Next() {
		var (
			path      string
			content   string
			embedding []byte
		)
		err := rows.Scan(&path, &content, &embedding)
		if err != nil {
			return "", err
		}

		contentEmb, err := decodeEmbedding(embedding)
		if err != nil {
			return "", err
		}
		score := cosineSimilarity(qEmb, contentEmb)
		scores = append(scores, scoreRecord{path, score, content})

		fmt.Printf("path: %s, score: %v, content: %d, embedding: %d\n", path, score, len(content), len(embedding))
		// fmt.Println(path, score)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	slices.SortFunc(scores, func(a, b scoreRecord) int {
		// The scores are in the range [0, 1], so scale them to get non-zero
		// integers for comparison.
		return int(100.0 * (a.Score - b.Score))
	})

	// Take the 3 best-scoring chunks
	var contextInfo string

	if len(scores) == 0 {
		fmt.Println("No scores found")
		return "", nil
	}

	cnt := 0
	for i := len(scores) - 1; i >= 0; i-- {
		contextInfo = contextInfo + "\n" + scores[i].Content
		cnt++
		if cnt >= max_scores {
			break
		}
	}
	return contextInfo, nil

}

// AnswerQuestion answers the given question by generating a response using OpenAI's chat completion API.
// It first retrieves the necessary rows from the database by calling the getRows() method.
// Then, it obtains the question's embedding by calling the GetEmbedding() method.
// Next, it retrieves the context information by calling the getContextInfo() method.
// It creates a chat completion request with the obtained context and the question, and sends it to the chat completion API.
// Finally, it prints the response and returns nil or an error if any occurred.
func (ol *RagollamaClient) AnswerQuestion(question1 string) error {

	rows, err := ol.getRows()
	if err != nil {
		return err
	}
	defer rows.Close()

	qEmb, err := ol.GetEmbedding(question1)
	if err != nil {
		return err
	}

	contextInfo, err := ol.getContextInfo(qEmb, rows, MAX_SCORES)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`Use the below information to answer the subsequent question.
Information:
%v

Question: %v`, contextInfo, question1)

	fmt.Println("--------------- QUERY -------------------\n", query, "\n-------------------------------------------\n")
	start := time.Now()

	resp, err := ol.CreateChatCompletion(
		openai.ChatCompletionRequest{
			Model: GetOllamaModel(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("Got response, ID:", resp.ID, "Duration:", time.Now().Sub(start))
	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	choice := resp.Choices[0]
	fmt.Println("Response.choices[0]:\n" + choice.Message.Content)
	return nil
}
