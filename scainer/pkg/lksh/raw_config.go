package lksh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"

	"github.com/antchfx/htmlquery"
)

type RawConfig struct {
	Title   string `json:"title"`
	BaseURL string `json:"base_url"`
	Layout  any    `json:"layout"`
}

func ExtractRawConfig(pageHTML []byte) (RawConfig, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(pageHTML))
	if err != nil {
		return RawConfig{}, fmt.Errorf("lksh: parse html: %w", err)
	}
	node := htmlquery.FindOne(doc, `//script[@x-data='raw_config']/text()`)
	if node == nil {
		return RawConfig{}, fmt.Errorf("lksh: raw_config script not found")
	}
	text := html.UnescapeString(htmlquery.InnerText(node))
	var cfg RawConfig
	if err := json.Unmarshal([]byte(text), &cfg); err != nil {
		return RawConfig{}, fmt.Errorf("lksh: raw_config json: %w", err)
	}
	return cfg, nil
}
