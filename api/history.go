package api

import (
	"time"
)

type History struct {
	data        []int
	size        int
	lastRequest time.Time
}

func newHistory(size int) *History {
	return &History{
		data:        make([]int, 0),
		size:        size,
		lastRequest: time.Now(),
	}
}

func (api *ApiConfig) AddToHistory(channelId string, context []int) {
	if _, ok := api.Channels[channelId]; !ok {
		api.Channels[channelId] = newHistory(api.HistoryTokensSize)
	}

	h := api.Channels[channelId]
	l := len(context)

	if l > h.size {
		h.data = context[l-h.size : l]
	} else {
		h.data = context
	}

	h.lastRequest = time.Now()
}

func (api *ApiConfig) DeleteOldHistories(delay time.Duration) {
	difference := time.Now().Add(delay)

	for channelId, history := range api.Channels {
		if history.lastRequest.Before(difference) {
			delete(api.Channels, channelId)
		}
	}
}

func (api *ApiConfig) GetHistory(channelId string) []int {
	if _, ok := api.Channels[channelId]; ok {
		return api.Channels[channelId].data
	}
	return nil
}

func (api *ApiConfig) DeleteChannelHistories(channelId string) bool {
	if _, ok := api.Channels[channelId]; !ok {
		return false
	}

	delete(api.Channels, channelId)
	return true
}
