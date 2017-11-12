package messagebus

import (
	"sync"
)

var (
	typeHandlers = make(map[string][]MessageHandler)
	mu           sync.Mutex
)

// Message is a message being transfered.
type Message struct {
	System string
	Type   string
	Data1  interface{}
	Data2  interface{}
}

// MessageHandler is any function which can be registered to receieve messages.
type MessageHandler func(m *Message)

// SendSync sends a message syncronously to any listener which is currently Registered to receive it.
func SendSync(m *Message) {
	mu.Lock()
	defer mu.Unlock()
	for _, h := range typeHandlers[m.Type] {
		h(m)
	}
}

// SendAsync sends a message asyncronously to any listener which is currently Registered to receive it.
func SendAsync(m *Message) {
	go SendSync(m)
}

// RegisterType registers a function to be called when a message is sent with matching type.
func RegisterType(t string, h MessageHandler) {
	mu.Lock()
	defer mu.Unlock()
	typeHandlers[t] = append(typeHandlers[t], h)
}
