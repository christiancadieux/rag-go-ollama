package ragollama

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	CREATE_EMBEDDINGS = `
         CREATE TABLE IF NOT EXISTS embeddings (
           id INTEGER PRIMARY KEY,
           embedding BLOB
		 )`
)

// CalculateEmbeddings calculates embeddings for all chunks in the database and stores them in the embeddings table.
func (ol *RagollamaClient) CalculateEmbeddings() error {

	db, err := sql.Open("sqlite3", ol.dbPath)
	if err != nil {
		return fmt.Errorf("sql.Open - %v", err)
	}
	defer db.Close()

	log.Println("Creating embeddings table if needed")
	_, err = db.Exec(CREATE_EMBEDDINGS)
	if err != nil {
		return fmt.Errorf("db.Exec - %v", err)
	}

	log.Println("Clearing embeddings table")
	_, err = db.Exec(`DELETE FROM embeddings`)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT * FROM chunks")
	if err != nil {
		return fmt.Errorf("db.Exec - %v", err)
	}
	defer rows.Close()

	// calculate embeddings for all chunks in the DB, storing them in embs.
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
		if err != nil {
			return fmt.Errorf("rows.Scan - %v", err)
		}
		fmt.Printf("id: %d, path: %s, nchunk: %d, content: %d\n", id, path, nchunk, len(content))
		if len(content) > 0 {
			emb0, err := ol.GetEmbedding(content)
			if err != nil {
				return err
			}
			emb, err := encodeEmbedding(emb0)
			if err != nil {
				return err
			}
			embs = append(embs, embData{id, emb})
		}
	}

	if err = rows.Err(); err != nil {
		return err
	}
	rows.Close()

	// Step 2: insert all embedding data into the embeddings table.
	for _, emb := range embs {
		fmt.Println("Inserting into embeddings, id", emb.id)
		_, err = db.Exec("INSERT INTO embeddings VALUES (?, ?)", emb.id, emb.data)
		if err != nil {
			return fmt.Errorf("insert embeddings - %v", err)
		}
	}
	return nil
}
