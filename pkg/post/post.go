package post

import (
	"context"
	"log"
	"time"

	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Downnvote = iota - 1
	Unvote
	Upvote
)

type PostRequest struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Title    string `json:"title,omitempty"`
	URL      string `json:"url,omitempty"`
	Type     string `json:"type"`
}

type Post struct {
	BSONID           primitive.ObjectID `json:"-" bson:"_id"`
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title,omitempty" bson:"title,omitempty"`
	URL              string             `json:"url,omitempty" bson:"url,omitempty"`
	Author           Author             `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	Text             string             `json:"text" bson:"text"`
	Votes            []*Vote            `json:"votes" bson:"votes"`
	Comments         []*Comment         `json:"comments" bson:"comments"`
	Created          time.Time          `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               string             `json:"id" bson:"id"`
}

type CommentsMap map[string]*Comment

// func (p *Post) MarshalJSON() ([]byte, error) {
// 	if len(p.ID) == 0 {
// 		p.ID = p.BSONID.String()
// 	}

// 	return json.Marshal(p)
// }

// func (p *Post) UnmarshalJSON(data []byte) error {
// 	_id, err := primitive.ObjectIDFromHex(p.ID)
// 	if err != nil {
// 		return err
// 	}

// 	p.BSONID = _id

// 	return json.Unmarshal(data, p)
// }

type Author struct {
	Username string `json:"username" bson:"username"`
	ID       string `json:"id" bson:"id"`
}

type Vote struct {
	ID     string             `json:"-" bson:"-"`
	BSONID primitive.ObjectID `json:"-" bson:"_id"`
	User   string             `json:"user" bson:"user"`
	Vote   int                `json:"vote" bson:"vote"`
}

type Comment struct {
	BSONID  primitive.ObjectID `json:"-" bson:"_id"`
	Created time.Time          `json:"created" bson:"created"`
	Author  *Author            `json:"author" bson:"author"`
	Body    string             `json:"body" bson:"body"`
	ID      string             `json:"id" bson:"id"`
}

type CommentRequest struct {
	Comment string `json:"comment"`
}

type PostRepo interface {
	List(ctx context.Context) (*[]Post, error)
	Create(ctx context.Context, user *user.User, pr *PostRequest) (*Post, error)
	Delete(ctx context.Context, post string) error
	Get(ctx context.Context, postID string) (*Post, error)
	GetByCategory(ctx context.Context, category string) (*[]Post, error)
	CreateComment(ctx context.Context, postId string, text string, user *user.User) (*Post, error)
	DeleteComment(ctx context.Context, postId string, commentID string, username string) (*Post, error)
	Vote(ctx context.Context, postId string, username string, voteVal int) (*Post, error)
	PostsByUser(ctx context.Context, username string) (*[]Post, error)
}

func NewPost(postRequest *PostRequest, user *user.User) *Post {
	vote := &Vote{
		User: user.ID,
		Vote: 1,
	}

	bsonId := primitive.NewObjectID()
	id := bsonId.Hex()
	log.Print(id)

	newPost := &Post{
		ID:     id,
		BSONID: bsonId,
		Type:   postRequest.Type,
		Title:  postRequest.Title,
		URL:    postRequest.URL,
		Author: Author{
			Username: user.Username,
			ID:       user.ID,
		},
		Category: postRequest.Category,
		Text:     postRequest.Text,
		Votes:    []*Vote{vote},
		Comments: make([]*Comment, 0),
		Created:  time.Now().UTC(),
	}
	newPost.CalcScoreAndUpvotePercentage()

	return newPost
}

func (p *Post) CalcScoreAndUpvotePercentage() {
	var votesAmount int
	var upvotesAmount int
	for _, vote := range p.Votes {
		voteVal := vote.Vote
		if voteVal == 1 {
			upvotesAmount += voteVal
		}
		votesAmount += voteVal
	}

	p.Score = votesAmount

	if votesAmount == 0 {
		p.UpvotePercentage = 0
	} else {
		p.UpvotePercentage = (upvotesAmount * 100) / len(p.Votes)
	}
}

func (p *Post) IncreaseViews() {
	p.Views++
}
