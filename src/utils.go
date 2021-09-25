package main

import (
	"encoding/json"
	"io"
	"sync"
)

func indexOf(arr []string, value string) int {
	for i, el := range arr {
		if el == value {
			return i
		}
	}
	return -1
}

func goLaunch(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

func readReadCloser(reader io.ReadCloser) []byte {
	if reader == nil {
		return nil
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}
	return data
}

func extractJson(byteData []byte) map[string]interface{} {
	if byteData == nil {
		return nil
	}
	var f map[string]interface{}
	err := json.Unmarshal(byteData, &f)
	if err != nil {
		return nil
	}
	return f
}
