package main

import "github.com/tiktoken-go/tokenizer"

func countTokens(text string) int {
	enc, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		panic(err)
	}
	ids, _, _ := enc.Encode(text)
	return len(ids)
}
