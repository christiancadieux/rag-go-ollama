
# DEFAULT ENV-VAR
# OLLAMA_URL=http://alien:11434
# OLLAMA_MODEL=mistral
# set export GO=go

all: build test

build:
	${GO} build -o rago ./cmd/rag/...


test: chunk calc q

chunk:
	./rago --chunk --db rag.db --clear 


calc:
	./rago --calculate --db rag.db

q:
	./rago --answer --db rag.db
