package gateway

import (
	"sort"
	"strings"
	"time"

	"github.com/DaylaHoang/entropix/pkg/api"
)

func assembleResponse(chunks []api.StreamChunk) *api.ChatCompletionResponse {
	if len(chunks) == 0 {
		return &api.ChatCompletionResponse{}
	}

	first := chunks[0]
	resp := &api.ChatCompletionResponse{
		ID:      first.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   first.Model,
	}

	choiceContents := make(map[int]*strings.Builder)
	var finishReason *string

	for _, chunk := range chunks {
		for _, choice := range chunk.Choices {
			if _, ok := choiceContents[choice.Index]; !ok {
				choiceContents[choice.Index] = &strings.Builder{}
			}
			if choice.Delta != nil {
				choiceContents[choice.Index].WriteString(choice.Delta.Content)
			}
			if choice.FinishReason != nil {
				fr := *choice.FinishReason
				finishReason = &fr
			}
		}
	}

	indices := make([]int, 0, len(choiceContents))
	for idx := range choiceContents {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	for _, idx := range indices {
		resp.Choices = append(resp.Choices, api.Choice{
			Index: idx,
			Message: &api.Message{
				Role:    "assistant",
				Content: choiceContents[idx].String(),
			},
			FinishReason: finishReason,
		})
	}

	return resp
}
