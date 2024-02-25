
# DEFAULT ENV-VAR
# OLLAMA_URL=http://alien:11434
# OLLAMA_MODEL=mistral

all: chunk calc q

chunk:
	go run ./cmd/chunker --outdb rag.db --clear 


calc:
	go run -v  ./cmd/rag --calculate --db rag.db

q:
	go run ./cmd/rag --answer --db rag.db
