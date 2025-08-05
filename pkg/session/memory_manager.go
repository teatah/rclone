package session

// import (
// 	"fmt"
// 	"sync"

// 	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/token"
// 	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
// )

// type JwtSessionManager struct {
// 	mu       *sync.RWMutex
// 	sessions map[string]*Session
// }

// func NewJwtSessionManager() *JwtSessionManager {
// 	return &JwtSessionManager{
// 		mu:       &sync.RWMutex{},
// 		sessions: make(map[string]*Session, 5),
// 	}
// }

// func (sm *JwtSessionManager) Create(user *user.User) (*Session, error) {
// 	sess, err := NewSession(user)

// 	if err != nil {
// 		return nil, err
// 	}

// 	sm.mu.Lock()
// 	sm.sessions[sess.ID] = sess
// 	sm.mu.Unlock()

// 	return sess, nil
// }

// func (sm *JwtSessionManager) Check(tokenString string) (*Session, error) {
// 	_, err := token.ParseJwt(tokenString)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sm.mu.Lock()
// 	sess, ok := sm.sessions[tokenString]
// 	if !ok {
// 		return nil, fmt.Errorf("session %s not found", tokenString)
// 	}
// 	sm.mu.Unlock()

// 	return sess, nil
// }
