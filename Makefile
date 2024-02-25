
# DEFAULT ENV-VAR
# OLLAMA_URL=http://alien:11434
# OLLAMA_MODEL=mistral
# set export GO=go

all: build test

build:
	${GO} build -o rago ./cmd/rag/...


test: chunk calc q

chunk:
	./rago --chunk  --clear  --rootdir ./docs --db rag.db


calc:
	./rago --calculate

q:
	./rago --answer

q2:
	export RAG_Q='What is RDEI?'; ./rago --answer 
