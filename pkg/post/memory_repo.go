package post

// import (
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// 	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
// )

// type PostMemoryRepo struct {
// 	mu    *sync.RWMutex
// 	posts map[string]*Post
// }

// func NewPostMemoryRepo() *PostMemoryRepo {
// 	return &PostMemoryRepo{
// 		mu:    &sync.RWMutex{},
// 		posts: make(map[string]*Post, 5),
// 	}
// }

// func (pr *PostMemoryRepo) List() []*Post {
// 	pr.mu.RLock()
// 	postsSlice := make([]*Post, 0, len(pr.posts))
// 	for _, post := range pr.posts {
// 		postsSlice = append(postsSlice, post)
// 	}
// 	pr.mu.RUnlock()

// 	return postsSlice
// }

// func (pr *PostMemoryRepo) Create(user *user.User, postRequest *PostRequest) *Post {
// 	newPost, postID := NewPost(postRequest, user)

// 	pr.mu.Lock()
// 	pr.posts[postID] = newPost
// 	pr.mu.Unlock()

// 	return newPost
// }

// func (pr *PostMemoryRepo) Delete(post *Post, username string) error {
// 	postCreator := post.Author.Username
// 	if postCreator != username {
// 		return fmt.Errorf("user %s can`t delete %s`s post", username, postCreator)
// 	}

// 	pr.mu.Lock()
// 	delete(pr.posts, post.ID)
// 	pr.mu.Unlock()

// 	return nil
// }

// func (pr *PostMemoryRepo) Get(postID string) (*Post, error) {
// 	pr.mu.Lock()
// 	post, ok := pr.posts[postID]
// 	pr.mu.Unlock()

// 	if !ok {
// 		return nil, fmt.Errorf("post with id %s not found", postID)
// 	}

// 	return post, nil
// }

// func (pr *PostMemoryRepo) GetByCategory(category string) []*Post {
// 	catPosts := make([]*Post, 0, 10)
// 	pr.mu.Lock()
// 	for _, post := range pr.posts {
// 		if post.Category == category {
// 			catPosts = append(catPosts, post)
// 		}
// 	}
// 	pr.mu.Unlock()

// 	return catPosts
// }

// func (pr *PostMemoryRepo) CreateComment(post *Post, text string, user *user.User) (*Post, error) {
// 	commentID := uuid.NewString()
// 	comment := &Comment{
// 		Created: time.Now().UTC(),
// 		Author: &Author{
// 			Username: user.Username,
// 			ID:       user.ID,
// 		},
// 		Body: text,
// 		ID:   commentID,
// 	}
// 	pr.mu.Lock()
// 	post.Comments[commentID] = comment
// 	pr.mu.Unlock()

// 	return post, nil
// }

// func (pr *PostMemoryRepo) DeleteComment(post *Post, commentID string, username string) (*Post, error) {
// 	pr.mu.Lock()
// 	comment, ok := post.Comments[commentID]
// 	if !ok {
// 		return nil, fmt.Errorf("no comment with id %s", commentID)
// 	}
// 	authorUsername := comment.Author.Username
// 	if authorUsername != username {
// 		return nil, fmt.Errorf("user %s can`t delete %s`s comment", authorUsername, username)
// 	}

// 	delete(post.Comments, commentID)
// 	pr.mu.Unlock()

// 	return post, nil
// }

// func (pr *PostMemoryRepo) Vote(post *Post, username string, voteVal int) *Post {
// 	pr.mu.Lock()
// 	if voteVal == 0 {
// 		delete(post.Votes, username)
// 	} else {
// 		vote := &Vote{
// 			User: username,
// 			Vote: voteVal,
// 		}

// 		post.Votes[username] = vote
// 	}
// 	post.CalcScoreAndUpvotePercentage()
// 	pr.mu.Unlock()

// 	return post
// }

// func (pr *PostMemoryRepo) PostsByUser(username string) []*Post {
// 	posts := make([]*Post, 0, len(pr.posts))

// 	pr.mu.Lock()
// 	for _, post := range pr.posts {
// 		if post.Author.Username == username {
// 			posts = append(posts, post)
// 		}
// 	}
// 	pr.mu.Unlock()

// 	return posts
// }
