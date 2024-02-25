package ragollama

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const (
	tokenEncoding = "cl100k_base"
	tableName     = "chunks"
	chunkSize     = 1000
	tableSql      = `CREATE TABLE IF NOT EXISTS %v (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  path TEXT,
	  nchunk INTEGER,
	  content TEXT
     );`
)

// Chunks breaks the files in the root directory into chunks and inserts them into the chunks table in the SQLite database.
// If clear is true, the chunks table will be cleared before inserting new chunks.
// Each chunk includes the file path, the chunk number, and the content of the chunk.
// An error is returned if there is any issue with the SQLite database or with the chunking process.
func (ol *RagollamaClient) Chunks(clear bool, rootDir string) error {

	db, err := sql.Open("sqlite3", ol.dbPath)
	if err != nil {
		return fmt.Errorf("sql.Open - %v", err)
	}

	_, err = db.Exec(fmt.Sprintf(tableSql, tableName))
	if err != nil {
		return fmt.Errorf("db.Exec - %v", err)
	}

	if clear {
		log.Printf("Clearing DB table %v", tableName)
		_, err := db.Exec(fmt.Sprintf("delete from %s", tableName))
		if err != nil {
			return fmt.Errorf("db.Exec - %v", err)
		}
	}

	insertStmt, err := db.Prepare("insert into chunks(path, nchunk, content) values(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("db.Prepare - %v", err)
	}

	tokTotal := 0
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".md" {
			log.Printf("Chunking %v", path)
			chunks, err := breakToChunks(path)
			if err != nil {
				return err
			}
			for i, chunk := range chunks {
				fmt.Println(path, i, len(chunk))
				tokTotal += len(chunk)
				_, err := insertStmt.Exec(path, i, chunk)
				if err != nil {
					return err
				}
			}

		}
		return nil
	})
	fmt.Println("Total tokens:", tokTotal)
	return nil
}

// breakToChunks reads the file in `path` and breaks it into chunks of
// approximately chunkSize tokens each, returning the chunks.
func breakToChunks(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	tke, err := tiktoken.GetEncoding(tokenEncoding)
	if err != nil {
		return nil, err
	}

	chunks := []string{""}

	scanner := bufio.NewScanner(f)
	scanner.Split(splitByParagraph)

	for scanner.Scan() {
		chunks[len(chunks)-1] = chunks[len(chunks)-1] + scanner.Text() + "\n"
		toks := tke.Encode(chunks[len(chunks)-1], nil, nil)
		if len(toks) > chunkSize {
			chunks = append(chunks, "")
		}
	}

	// If we added a new empty chunk but there weren't any paragraphs to add to
	// it, make sure to remove it.
	if len(chunks[len(chunks)-1]) == 0 {
		chunks = chunks[:len(chunks)-1]
	}

	return chunks, nil
}

// splitByParagraph is a custom split function for bufio.Scanner to split by
// paragraphs (text pieces separated by two newlines).
func splitByParagraph(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, bytes.TrimSpace(data[:i]), nil
	}

	if atEOF && len(data) != 0 {
		return len(data), bytes.TrimSpace(data), nil
	}

	return 0, nil, nil
}
