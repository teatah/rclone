package post

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostMemoryRepo struct {
	mu        *sync.RWMutex
	posts     map[string]*Post
	postsColl *mongo.Collection
}

func NewPostMemoryRepo(postsCollection *mongo.Collection) *PostMemoryRepo {
	return &PostMemoryRepo{
		mu:        &sync.RWMutex{},
		posts:     make(map[string]*Post, 5),
		postsColl: postsCollection,
	}
}

func (pr *PostMemoryRepo) List(ctx context.Context) (*[]Post, error) {
	var allPosts []Post

	cur, err := pr.postsColl.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &allPosts)

	return &allPosts, err
}

func (pr *PostMemoryRepo) Create(ctx context.Context, user *user.User, postRequest *PostRequest) (*Post, error) {
	newPost := NewPost(postRequest, user)

	_, err := pr.postsColl.InsertOne(ctx, newPost)
	if err != nil {
		return nil, err
	}

	return newPost, nil
}

func (pr *PostMemoryRepo) Delete(ctx context.Context, postId string) error {
	bsonId, err := primitive.ObjectIDFromHex(postId)
	if err != nil {
		return err
	}

	res, err := pr.postsColl.DeleteOne(ctx, bson.M{"_id": bsonId})
	if err == nil && res.DeletedCount == 0 {
		err = fmt.Errorf("failed ro delete post %s: post not found", postId)
	}

	return err
}

func (pr *PostMemoryRepo) Get(ctx context.Context, postID string) (*Post, error) {
	_id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	post := &Post{}

	filter := bson.M{"_id": _id}
	update := bson.M{"$inc": bson.M{"views": 1}}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = pr.postsColl.FindOneAndUpdate(ctx, filter, update, opt).Decode(post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("post with id %s not found", postID)
		}

		return nil, err
	}

	return post, nil
}

func (pr *PostMemoryRepo) GetByCategory(ctx context.Context, category string) (*[]Post, error) {
	return pr.listByKeyValue(ctx, "category", category)
}

func (pr *PostMemoryRepo) CreateComment(
	ctx context.Context,
	postId string,
	text string,
	user *user.User,
) (*Post, error) {
	commentBSONID := primitive.NewObjectID()
	comment := &Comment{
		BSONID:  commentBSONID,
		ID:      commentBSONID.Hex(),
		Created: time.Now().UTC(),
		Author: &Author{
			Username: user.Username,
			ID:       user.ID,
		},
		Body: text,
	}

	updatedPost := &Post{}

	bsonId, err := primitive.ObjectIDFromHex(postId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonId}
	update := bson.M{"$push": bson.M{"comments": comment}}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = pr.postsColl.FindOneAndUpdate(ctx, filter, update, opt).Decode(updatedPost)

	return updatedPost, err
}

func (pr *PostMemoryRepo) DeleteComment(ctx context.Context, postId string, commentID string, username string) (*Post, error) {
	commentBSONID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return nil, err
	}

	updatedPost := &Post{}

	bsonId, err := primitive.ObjectIDFromHex(postId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonId}
	update := bson.M{"$pull": bson.M{"comments": bson.M{"_id": commentBSONID}}}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = pr.postsColl.FindOneAndUpdate(ctx, filter, update, opt).Decode(updatedPost)

	return updatedPost, err
}

func (pr *PostMemoryRepo) Vote(ctx context.Context, postId string, username string, voteVal int) (*Post, error) {
	bsonId, err := primitive.ObjectIDFromHex(postId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonId}

	pullUpdate := bson.M{"$pull": bson.M{"votes": bson.M{"user": username}}}

	_, err = pr.postsColl.UpdateOne(ctx, filter, pullUpdate)
	if err != nil {
		return nil, err
	}

	updatedPost := &Post{}

	err = pr.postsColl.FindOne(ctx, filter).Decode(updatedPost)
	if err != nil {
		return nil, err
	}

	if voteVal != 0 {
		vote := &Vote{
			User: username,
			Vote: voteVal,
		}

		updatedPost.Votes = append(updatedPost.Votes, vote)
	}
	updatedPost.CalcScoreAndUpvotePercentage()

	updateStats := bson.M{
		"$set": bson.M{
			"score":            updatedPost.Score,
			"upvotePercentage": updatedPost.UpvotePercentage,
			"votes":            updatedPost.Votes,
		},
	}

	_, err = pr.postsColl.UpdateOne(ctx, filter, updateStats)

	return updatedPost, err
}

func (pr *PostMemoryRepo) PostsByUser(ctx context.Context, username string) (*[]Post, error) {
	return pr.listByKeyValue(ctx, "author.username", username)
}

func (pr *PostMemoryRepo) listByKeyValue(ctx context.Context, key, value string) (*[]Post, error) {
	var catPosts []Post
	cur, err := pr.postsColl.Find(ctx, bson.M{key: value})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &catPosts)

	return &catPosts, err
}
