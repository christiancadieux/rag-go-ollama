
# DEFAULT ENV-VAR
# OLLAMA_URL=http://alien:11434
# OLLAMA_MODEL=mistral
# set export GO=go

all: chunk calc q

chunk:
	${GO} run ./cmd/chunker --outdb rag.db --clear 


calc:
	${GO} run -v  ./cmd/rag --calculate --db rag.db

q:
	${GO} run ./cmd/rag --answer --db rag.db
