package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const OLLAMA_URL = "http://alien:11434"
const OLLAMA_MODEL = "mistral"

type OllamaClient struct {
	config openai.ClientConfig
	client *openai.Client
	ctx    context.Context
}

func GetOllamaUrl() string {
	ollama_url := os.Getenv("OLLAMA_URL")
	if ollama_url == "" {
		ollama_url = OLLAMA_URL
	}
	return ollama_url
}

func GetOllamaModel() string {
	ollama_model := os.Getenv("OLLAMA_MODEL")
	if ollama_model == "" {
		ollama_model = OLLAMA_MODEL
	}
	return ollama_model
}

func NewOllamaClient() *OllamaClient {
	cl, cfg := NewClientWithBase(GetOllamaUrl())
	return &OllamaClient{client: cl, config: cfg, ctx: context.Background()}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// getEmbedding invokes the OpenAI embedding API to calculate the embedding
// for the given string. It returns the embedding.
func (o *OllamaClient) GetEmbedding(data string) []float32 {

	queryResponse, err := o.CreateEmbeddingsOllama(GetOllamaUrl(), GetOllamaModel(), data)

	checkErr(err)
	return queryResponse.Data[0].Embedding
}

func NewClientWithBase(base string) (*openai.Client, openai.ClientConfig) {
	config := DefaultConfigWithBase(base)
	return openai.NewClientWithConfig(config), config
}

func DefaultConfigWithBase(base string) openai.ClientConfig {
	thisClient := &http.Client{}

	return openai.ClientConfig{
		BaseURL:            base,
		APIType:            openai.APITypeOpenAI,
		OrgID:              "",
		HTTPClient:         thisClient,
		EmptyMessagesLimit: 10,
	}
}

type ollama struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func (o *OllamaClient) CreateEmbeddingsOllama(baseurl, llmModel string, data string) (*openai.EmbeddingResponse, error) {

	url2 := baseurl + "/api/embeddings"
	ol := ollama{}
	ol.Model = llmModel
	ol.Prompt = data

	data2, err := json.Marshal(ol)

	req, err := http.NewRequestWithContext(o.ctx, "POST", url2, strings.NewReader(string(data2)))

	if err != nil {
		fmt.Printf("NewRequest error %v \n", err)
		return nil, err
	}

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("http Do err", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println("BODY=", string(body))
	res := openai.EmbeddingResponse{}
	res.Data = []openai.Embedding{}
	res.Data = append(res.Data, openai.Embedding{})

	json.Unmarshal(body, &res.Data[0])

	err = o.sendRequest(req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type OllamaCompletionRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream,omitempty"`
	Prompt string `json:"prompt"`
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (o *OllamaClient) CreateChatCompletion(request openai.ChatCompletionRequest) (response openai.ChatCompletionResponse, err error) {
	if request.Stream {
		err = openai.ErrChatCompletionStreamNotSupported
		return
	}
	oRequest := OllamaCompletionRequest{}
	oRequest.Model = request.Model
	oRequest.Stream = false
	oRequest.Prompt = request.Messages[0].Content

	urlSuffix := "/v1/chat/completions"

	req, err := o.newRequest(http.MethodPost, o.config.BaseURL+urlSuffix, withBody(request))
	if err != nil {
		return
	}

	err = o.sendRequest(req, &response)
	return
}
