package chat

import "sync"

const subscriberBufferSize = 64

type EventBus struct {
	subscribers sync.Map // map[string]map[string]chan Event — sessionID → subscriberID → channel
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (b *EventBus) Subscribe(sessionID, subscriberID string) <-chan Event {
	ch := make(chan Event, subscriberBufferSize)

	val, _ := b.subscribers.LoadOrStore(sessionID, &sync.Map{})
	subs := val.(*sync.Map)
	subs.Store(subscriberID, ch)

	return ch
}

func (b *EventBus) Unsubscribe(sessionID, subscriberID string) {
	val, ok := b.subscribers.Load(sessionID)
	if !ok {
		return
	}
	subs := val.(*sync.Map)
	if chVal, ok := subs.LoadAndDelete(subscriberID); ok {
		close(chVal.(chan Event))
	}
}

func (b *EventBus) Publish(event Event) {
	val, ok := b.subscribers.Load(event.SessionID)
	if !ok {
		return
	}
	subs := val.(*sync.Map)
	subs.Range(func(_, chVal any) bool {
		ch := chVal.(chan Event)
		select {
		case ch <- event:
		default:
			// drop event if subscriber buffer is full
		}
		return true
	})
}
