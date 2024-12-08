package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type Request struct {
	Question string `json:"question"`
}

type Response struct {
	Answer string `json:"answer"`
}

/*type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

*/

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./index.html")
	})

	mux.HandleFunc("/fetva", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		answer, err := getOpenAIResponse(req.Question)
		if err != nil {
			http.Error(w, "Failed to get response from OpenAI", http.StatusInternalServerError)
			return
		}

		resp := Response{Answer: answer}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("Server is running on localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}

func getOpenAIResponse(question string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"
	apiKey := "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"

	data := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": "Sen bir fetva hocasısın. Soruları İslami açıdan yanıtla. Cevapların kısa ve net olsun"},
			{"role": "user", "content": question},
		},
		"max_tokens":  150,
		"temperature": 0.7,
	}
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "No response from OpenAI", nil
}
