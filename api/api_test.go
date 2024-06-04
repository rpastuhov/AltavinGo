package api

import (
	"fmt"
	"testing"
)

var api = &ApiConfig{
	ApiDomain: "http://localhost:11434",
	Model:     "llama3:8b",
	Channels:  make(map[string]*History),
}

func TestGenerage(t *testing.T) {

	var arr []int
	for i := 0; i < 99; i++ {
		arr = append(arr, i)
	}

	fmt.Println(arr)

	api.AddToHistory("sgs", arr)
	api.AddToHistory("sgs", api.Channels["sgs"].data)

	// fmt.Println(api.Channels["sgs"].data)

}
