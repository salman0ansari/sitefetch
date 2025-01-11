package main

import (
	"net/url"

	"github.com/tiktoken-go/tokenizer"
)

func countTokens(text string) int {
	enc, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		panic(err)
	}
	ids, _, _ := enc.Encode(text)
	return len(ids)
}

func normalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Path == "/" {
		u.Path = ""
	}
	return u.Scheme + "://" + u.Host + u.Path
}
