package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	client          = &http.Client{}
	GptModel        = os.Getenv("GPT_MODEL")
	GptMaxTokens, _ = strconv.Atoi(os.Getenv("GPT_MAX_TOKENS"))
	GptURL          = os.Getenv("GPT_URL")
	GptToken        = os.Getenv("GPT_TOKEN")
)

func Propmt(w http.ResponseWriter, r *http.Request) {
	// verify method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, `{"error": "method %s not allowed"}`, r.Method)
		return
	}

	// read body
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var prompt *Prompt

	if err := json.Unmarshal(body, &prompt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	// logic

	msgs := []Message{
		{
			Role:    "system",
			Content: prompt.Context,
		},
		{
			Role:    "user",
			Content: prompt.Message,
		},
	}
	gptRes, err := SendRequest(msgs)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	res := map[string]interface{}{
		"prompt_result": gptRes.Choices[len(gptRes.Choices)-1].Message.Content,
		"finish_reason": gptRes.Choices[len(gptRes.Choices)-1].FinishReason,
		"model":         gptRes.Model,
		"used_tokens":   gptRes.Usage.TotalTokens,
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

func SendRequest(msgs []Message) (*GptResponse, error) {
	const op = "SendRequest"
	gptRequest := &GptRequest{
		Model:       GptModel,
		Messages:    msgs,
		Temperature: 0.2,
		MaxTokens:   GptMaxTokens,
	}

	b, err := json.Marshal(gptRequest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req, err := http.NewRequest("POST", GptURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// fmt.Println("request: ", string(b))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.Trim(GptToken, "\"")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w, %s", op, err, resp.Status)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	res := &GptResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return res, nil
}

type GptResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
type GptRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Prompt struct {
	Context string `json:"context"`
	Message string `json:"message"`
}
