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

func (ol *RagollamaClient) CalculateEmbeddings() {

	db, err := sql.Open("sqlite3", ol.dbPath)
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
