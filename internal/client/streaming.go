package client

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ollama/ollama/api"
)

// StreamTokenMsg represents a single token received from the stream
type StreamTokenMsg struct {
	Token     string
	Timestamp time.Time
}

// StreamDoneMsg indicates the stream has completed successfully
type StreamDoneMsg struct {
	TotalTokens int
	Duration    time.Duration
}

// StreamErrorMsg indicates the stream encountered an error
type StreamErrorMsg struct {
	Err           error
	PartialTokens int
}

// StreamHandler manages real-time token streaming using Bubble Tea commands
type StreamHandler struct {
	client    *api.Client
	modelName string
	msgChan   chan tea.Msg
	mu        sync.Mutex
}

// NewStreamHandler creates a new StreamHandler
func NewStreamHandler(client *api.Client, modelName string) *StreamHandler {
	return &StreamHandler{
		client:    client,
		modelName: modelName,
	}
}

// Stream initiates a streaming chat request
func (h *StreamHandler) Stream(ctx context.Context, messages []api.Message) tea.Cmd {
	// Create a new channel for this stream
	h.mu.Lock()
	h.msgChan = make(chan tea.Msg, 100)
	msgChan := h.msgChan
	h.mu.Unlock()

	// Start the streaming goroutine
	go func() {
		defer close(msgChan)

		startTime := time.Now()
		tokenCount := 0

		err := h.client.Chat(ctx, &api.ChatRequest{
			Model:    h.modelName,
			Messages: messages,
		}, func(resp api.ChatResponse) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Send each token as it arrives
			if resp.Message.Content != "" {
				tokenCount++
				select {
				case msgChan <- StreamTokenMsg{
					Token:     resp.Message.Content,
					Timestamp: time.Now(),
				}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})

		// Send completion or error message
		if err != nil {
			msgChan <- StreamErrorMsg{
				Err:           err,
				PartialTokens: tokenCount,
			}
		} else {
			msgChan <- StreamDoneMsg{
				TotalTokens: tokenCount,
				Duration:    time.Since(startTime),
			}
		}
	}()

	// Return a command that reads from the channel
	return h.waitForMsg()
}

// waitForMsg returns a command that waits for the next message from the stream
func (h *StreamHandler) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		h.mu.Lock()
		msgChan := h.msgChan
		h.mu.Unlock()

		if msgChan == nil {
			return nil
		}

		msg, ok := <-msgChan
		if !ok {
			return nil
		}
		return msg
	}
}

// WaitForNextMsg returns a command to continue receiving stream messages
func (h *StreamHandler) WaitForNextMsg() tea.Cmd {
	return h.waitForMsg()
}
