package user

// import (
// 	"errors"
// 	"sync"

// 	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/responses"
// )

// var (
// 	UserAlreadyExists = "already exists"
// 	UserNotFound      = "user not found"
// 	InvalidPassword   = "invalid password"
// )

// type UserMemoryRepo struct {
// 	mu   *sync.RWMutex
// 	data map[string]*User
// }

// func NewUserMemoryRepo() *UserMemoryRepo {
// 	return &UserMemoryRepo{
// 		mu:   &sync.RWMutex{},
// 		data: make(map[string]*User, 0),
// 	}
// }

// func (ur *UserMemoryRepo) Register(userRequest *UserRequest) (*User, error) {
// 	username := userRequest.Username
// 	password := userRequest.Password

// 	ur.mu.Lock()
// 	if _, ok := ur.data[username]; ok {
// 		err := responses.NewResponseError("body", "username", "username", UserAlreadyExists)

// 		return nil, err
// 	}

// 	newUser, err := NewUserWithCredentials(username, password)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ur.data[username] = newUser
// 	ur.mu.Unlock()

// 	return newUser, nil
// }

// func (ur *UserMemoryRepo) Login(userRequest *UserRequest) (*User, error) {
// 	ur.mu.RLock()
// 	user, ok := ur.data[userRequest.Username]
// 	if !ok {
// 		return nil, errors.New(UserNotFound)
// 	}

// 	err := user.CheckPassword(userRequest.Password)
// 	if err != nil {
// 		return nil, errors.New(InvalidPassword)
// 	}
// 	ur.mu.RUnlock()

// 	return user, nil
// }

// func (ur *UserMemoryRepo) GetUserByName(username string) (*User, error) {
// 	ur.mu.Lock()
// 	user, ok := ur.data[username]
// 	ur.mu.Unlock()

// 	if !ok {
// 		return nil, errors.New(UserNotFound)
// 	}

// 	return user, nil
// }
