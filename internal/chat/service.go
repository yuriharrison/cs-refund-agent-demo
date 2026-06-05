package chat

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu       sync.RWMutex
	sessions map[string]*session
	EventBus *EventBus
}

type session struct {
	ID        string
	Messages  []Message
	Escalated bool
}

func NewService(bus *EventBus) *Service {
	return &Service{
		sessions: make(map[string]*session),
		EventBus: bus,
	}
}

func (s *Service) CreateSession() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()
	s.sessions[id] = &session{ID: id}
	return id
}

func (s *Service) GetOrCreateSession(sessionID string) string {
	if sessionID != "" {
		s.mu.RLock()
		_, exists := s.sessions[sessionID]
		s.mu.RUnlock()
		if exists {
			return sessionID
		}
	}
	return s.CreateSession()
}

func (s *Service) AddMessage(sessionID string, role Role, content string) Message {
	msg := Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		sess = &session{ID: sessionID}
		s.sessions[sessionID] = sess
	}
	sess.Messages = append(sess.Messages, msg)

	return msg
}

func (s *Service) GetHistory(sessionID string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}

	msgs := make([]Message, len(sess.Messages))
	copy(msgs, sess.Messages)
	return msgs
}

func (s *Service) MarkEscalated(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[sessionID]; ok {
		sess.Escalated = true
	}
}

func (s *Service) IsEscalated(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, ok := s.sessions[sessionID]; ok {
		return sess.Escalated
	}
	return false
}

func (s *Service) Reset(sessionID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	newID := uuid.New().String()
	s.sessions[newID] = &session{ID: newID}
	return newID
}
