package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type streamChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
}

type streamPayload struct {
	Choices []streamChoice `json:"choices"`
}

func ParseChatCompletionSSE(r io.Reader, onDelta func(string) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventData []string
	flush := func() error {
		if len(eventData) == 0 {
			return nil
		}
		payload := strings.TrimSpace(strings.Join(eventData, "\n"))
		eventData = eventData[:0]
		if payload == "" || payload == "[DONE]" {
			return nil
		}
		var chunk streamPayload
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("openai: parse stream chunk: %w", err)
		}
		if len(chunk.Choices) == 0 {
			return nil
		}
		text := chunk.Choices[0].Delta.Content
		if text == "" {
			return nil
		}
		return onDelta(text)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			eventData = append(eventData, strings.TrimSpace(line[len("data:"):]))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

func FormatChatCompletionSSE(deltas ...string) []byte {
	var buf bytes.Buffer
	for _, d := range deltas {
		payload, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{
				{"delta": map[string]string{"content": d}},
			},
		})
		buf.WriteString("data: ")
		buf.Write(payload)
		buf.WriteString("\n\n")
	}
	buf.WriteString("data: [DONE]\n\n")
	return buf.Bytes()
}
