
## RaGoLLama


Golang Retrieval-Augmented Generation (RAG) application using ollama:

   - Pure GO implementation.
   - Using ollama with O/S LLMs running locally (Mistral, Gemma etc..).
   - Using github.com/sashabaranov/go-openai.
   - Based on github.com/eliben/code-for-blog/2023/go-rag-openai - updated to use ollama.

https://eli.thegreenplace.net/2023/retrieval-augmented-generation-in-go/

![Figure-5-RAG_Architecture max-800x800](https://github.com/christiancadieux/rag-go-ollama/assets/10535265/18dbca63-46d2-4e90-8528-c08b114f226d)
   
## OLLAMA

Start ollama on a server with GPU:

```
$ export OLLAMA_HOST=0.0.0.0:11434
$ ollama serve

# load and test model
$ ollama run mistral

# run ollama on a different server
$ export OLLAMA_HOST=[your-ollama-server]:11434
$ ollama run mistral
...

```

## USAGE

```
# ENV-VAR
export OLLAMA_URL=http://[your-ollama-server|localhost]:11434
export OLLAMA_MODEL="mistral"    # default

Run the chunker to populate the initial chunks DB:

$ go run ./cmd/rag --db rag.db --clear 

Calculate embeddings and store in DB:

$ go run ./cmd/rag --calculate --db rag.db

Ask Question:

$ go run ./cmd/rag --answer --db rag.db

```


## SQLITE3

Exploring the chunks DB from the command-line:

```
$ sqlite3 rag.db

> .tables
> select id, path, nchunk, length(content) from chunks;
> select id, length(embedding) from embeddings

```


## Sample Run - default model: Mistral

```
$ export GO=go; make

go build -o rago ./cmd/rag/...
./rago --chunk --db rag.db --clear 

2024/02/24 20:14:45 Clearing DB table chunks
2024/02/24 20:14:45 Chunking docs/rdei.md
docs/rdei.md 0 146
2024/02/24 20:14:45 Chunking docs/rdei2.md
docs/rdei2.md 0 217
2024/02/24 20:14:45 Chunking docs/rdei3.md
docs/rdei3.md 0 416
2024/02/24 20:14:45 Chunking docs/rdei4.md
docs/rdei4.md 0 342
Total tokens: 1121

./rago  --calculate --db rag.db
github.com/christiancadieux/rag-go-ollama/pkg/ollama
github.com/christiancadieux/rag-go-ollama/cmd/rag
Using LLM: http://alien:11434
2024/02/24 20:14:46 Creating embeddings table if needed
2024/02/24 20:14:46 Clearing embeddings table
id: 194, path: docs/rdei.md, nchunk: 0, content: 146
id: 195, path: docs/rdei2.md, nchunk: 0, content: 217
id: 196, path: docs/rdei3.md, nchunk: 0, content: 416
id: 197, path: docs/rdei4.md, nchunk: 0, content: 342
Inserting into embeddings, id 194
Inserting into embeddings, id 195
Inserting into embeddings, id 196
Inserting into embeddings, id 197

./rago --answer --db rag.db
Using LLM: http://alien:11434
path: docs/rdei.md, content: 146, embedding: 16384
docs/rdei.md 0.107603274
path: docs/rdei2.md, content: 217, embedding: 16384
docs/rdei2.md 0.12229379
path: docs/rdei3.md, content: 416, embedding: 16384
docs/rdei3.md 0.12253108
path: docs/rdei4.md, content: 342, embedding: 16384
docs/rdei4.md 0.0854641

============== QUERY =====================
 Use the below information to answer the subsequent question.
Information:

Computer science is the study of computation, information, and automation.[1][2][3] Computer science spans theoretical disciplines (such as algorithms, theory of computation, and information theory) to applied disciplines (including the design and implementation of hardware and software).[4][5][6] Though more often considered an academic discipline, computer science is closely related to computer programming.[7]

RDEI teams are a logical and resource divider for groups of users within RDEI. Teams let you organize users as well as discern between who can and can not create resources or tokens for the team. See [ Access Guide ]

For more information about Stork and Static Persistent Block Storage(PBS) volume setup, see  [ Creating Storage ] and [ Creating Public Storage ]

Algorithms and data structures are central to computer science.[8] The theory of computation concerns abstract models of computation and general classes of problems that can be solved using them. The fields of cryptography and computer security involve studying the means for secure communication and for preventing security vulnerabilities.


Question: what is a RDEI team? 
===================================

Got response, ID: chatcmpl-101 Duration: 801.793674ms
{
  "id": "chatcmpl-101",
  "object": "chat.completion",
  "created": 1708830890,
  "model": "mistral",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": " A RDEI team is a logical and resource divider for groups of users within the RDEI (Rackspace Data Execution Interface) platform. It lets you organize users and discern between who can and cannot create resources or tokens for the team, functioning as a way to manage access and permissions."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 264,
    "completion_tokens": 65,
    "total_tokens": 329
  }
}

Response.choices[0]:
 A RDEI team is a logical and resource divider for groups of users within the RDEI (Rackspace Data Execution Interface) platform. It lets you organize users and discern between who can and cannot create resources or tokens for the team, functioning as a way to manage access and permissions.

```

## Sample Run using model: Gemma

```
$ export OLLAMA_MODEL=gemma
# make sure you have loaded gemma with `ollama pull gemma` on your LLM server.
$ make test
...

Question: what is a RDEI team? 
===================================

Got response, ID: chatcmpl-261 Duration: 762.522581ms
{
  "id": "chatcmpl-261",
  "object": "chat.completion",
  "created": 1708831799,
  "model": "gemma",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Sure, here is the answer:\n\nRDEI teams are logical and resource dividers for groups of users within RDEI. They let you organize users as well as discern between who can and can not create resources or tokens for the team."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 252,
    "completion_tokens": 50,
    "total_tokens": 302
  }
}
Response.choices[0]:
Sure, here is the answer:

RDEI teams are logical and resource dividers for groups of users within RDEI. They let you organize users as well as discern between who can and can not create resources or tokens for the team.


```


## OLLAMA LOG: loading/using gemma

```
Feb 24 20:27:27: llama_model_loader: - type  f32:   57 tensors
Feb 24 20:27:27: llama_model_loader: - type q4_0:  196 tensors
Feb 24 20:27:27: llama_model_loader: - type q8_0:    1 tensors
Feb 24 20:27:27: llm_load_vocab: mismatch in special tokens definition ( 416/256000 vs 260/256000 ).
Feb 24 20:27:27: llm_load_print_meta: format           = GGUF V3 (latest)
Feb 24 20:27:27: llm_load_print_meta: arch             = gemma
Feb 24 20:27:27: llm_load_print_meta: vocab type       = SPM
Feb 24 20:27:27: llm_load_print_meta: n_vocab          = 256000
Feb 24 20:27:27: llm_load_print_meta: n_merges         = 0
Feb 24 20:27:27: llm_load_print_meta: n_ctx_train      = 8192
Feb 24 20:27:27: llm_load_print_meta: n_embd           = 3072
Feb 24 20:27:27: llm_load_print_meta: n_head           = 16
Feb 24 20:27:27: llm_load_print_meta: n_head_kv        = 16
Feb 24 20:27:27: llm_load_print_meta: n_layer          = 28
Feb 24 20:27:27: llm_load_print_meta: n_rot            = 192
Feb 24 20:27:27: llm_load_print_meta: n_embd_head_k    = 256
Feb 24 20:27:27: llm_load_print_meta: n_embd_head_v    = 256
Feb 24 20:27:27: llm_load_print_meta: n_gqa            = 1
Feb 24 20:27:27: llm_load_print_meta: n_embd_k_gqa     = 4096
Feb 24 20:27:27: llm_load_print_meta: n_embd_v_gqa     = 4096
Feb 24 20:27:27: llm_load_print_meta: f_norm_eps       = 0.0e+00
Feb 24 20:27:27: llm_load_print_meta: f_norm_rms_eps   = 1.0e-06
Feb 24 20:27:27: llm_load_print_meta: f_clamp_kqv      = 0.0e+00
Feb 24 20:27:27: llm_load_print_meta: f_max_alibi_bias = 0.0e+00
Feb 24 20:27:27: llm_load_print_meta: n_ff             = 24576
Feb 24 20:27:27: llm_load_print_meta: n_expert         = 0
Feb 24 20:27:27: llm_load_print_meta: n_expert_used    = 0
Feb 24 20:27:27: llm_load_print_meta: rope scaling     = linear
Feb 24 20:27:27: llm_load_print_meta: freq_base_train  = 10000.0
Feb 24 20:27:27: llm_load_print_meta: freq_scale_train = 1
Feb 24 20:27:27: llm_load_print_meta: n_yarn_orig_ctx  = 8192
Feb 24 20:27:27: llm_load_print_meta: rope_finetuned   = unknown
Feb 24 20:27:27: llm_load_print_meta: model type       = 7B
Feb 24 20:27:27: llm_load_print_meta: model ftype      = Q4_0
Feb 24 20:27:27: llm_load_print_meta: model params     = 8.54 B
Feb 24 20:27:27: llm_load_print_meta: model size       = 4.84 GiB (4.87 BPW)
Feb 24 20:27:27: llm_load_print_meta: general.name     = gemma-7b-it
Feb 24 20:27:27: llm_load_print_meta: BOS token        = 2 '<bos>'
Feb 24 20:27:27: llm_load_print_meta: EOS token        = 1 '<eos>'
Feb 24 20:27:27: llm_load_print_meta: UNK token        = 3 '<unk>'
Feb 24 20:27:27: llm_load_print_meta: PAD token        = 0 '<pad>'
Feb 24 20:27:27: llm_load_print_meta: LF token         = 227 '<0x0A>'
Feb 24 20:27:27: llm_load_tensors: ggml ctx size =    0.19 MiB
Feb 24 20:27:28: llm_load_tensors: offloading 28 repeating layers to GPU
Feb 24 20:27:28: llm_load_tensors: offloading non-repeating layers to GPU
Feb 24 20:27:28: llm_load_tensors: offloaded 29/29 layers to GPU
Feb 24 20:27:28: llm_load_tensors:        CPU buffer size =   796.88 MiB
Feb 24 20:27:28: llm_load_tensors:      CUDA0 buffer size =  4955.54 MiB
Feb 24 20:27:29: ...........................................................................
Feb 24 20:27:29: llama_new_context_with_model: n_ctx      = 2048
Feb 24 20:27:29: llama_new_context_with_model: freq_base  = 10000.0
Feb 24 20:27:29: llama_new_context_with_model: freq_scale = 1
Feb 24 20:27:29: ggml_init_cublas: GGML_CUDA_FORCE_MMQ:   yes
Feb 24 20:27:29: ggml_init_cublas: CUDA_USE_TENSOR_CORES: no
Feb 24 20:27:29: ggml_init_cublas: found 1 CUDA devices:
Feb 24 20:27:29:   Device 0: NVIDIA GeForce RTX 3080, compute capability 8.6, VMM: yes
Feb 24 20:27:29: llama_kv_cache_init:      CUDA0 KV buffer size =   896.00 MiB
Feb 24 20:27:29: llama_new_context_with_model: KV self size  =  896.00 MiB, K (f16):  448.00 MiB, V (f16):  448.00 MiB
Feb 24 20:27:29: llama_new_context_with_model:  CUDA_Host input buffer size   =    11.02 MiB
Feb 24 20:27:29: llama_new_context_with_model:      CUDA0 compute buffer size =   506.00 MiB
Feb 24 20:27:29: llama_new_context_with_model:  CUDA_Host compute buffer size =     6.00 MiB
Feb 24 20:27:29: llama_new_context_with_model: graph splits (measure): 3
Feb 24 20:27:29: time=2024-02-24T20:27:29.236-07:00 level=INFO source=dyn_ext_server.go:161 msg="Starting llama main loop"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |  3.948269702s |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |   97.863419ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |   99.882903ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |  101.099916ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |  117.700645ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |  117.487306ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:29: [GIN] 2024/02/24 - 20:27:29 | 200 |   99.514093ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:30: [GIN] 2024/02/24 - 20:27:30 | 200 |    99.77478ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:30: [GIN] 2024/02/24 - 20:27:30 | 200 |    97.42509ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:30: [GIN] 2024/02/24 - 20:27:30 | 200 |   95.765198ms |    192.168.0.51 | POST     "/api/embeddings"
Feb 24 20:27:31: [GIN] 2024/02/24 - 20:27:31 | 200 |  790.455482ms |    192.168.0.51 | POST     "/v1/chat/completions"

```


