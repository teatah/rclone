package post

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teatah/rclone/pkg/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostDBRepo struct {
	postsColl *mongo.Collection
}

func NewPostDBRepo(postsCollection *mongo.Collection) *PostDBRepo {
	return &PostDBRepo{
		postsColl: postsCollection,
	}
}

func (pr *PostDBRepo) AllPosts(ctx context.Context) (*[]Post, error) {
	var allPosts []Post

	cur, err := pr.postsColl.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &allPosts)

	return &allPosts, err
}

func (pr *PostDBRepo) CreatePost(ctx context.Context, user *user.User, postRequest *PostRequest) (*Post, error) {
	newPost := NewPost(postRequest, user)

	_, err := pr.postsColl.InsertOne(ctx, newPost)
	if err != nil {
		return nil, err
	}

	return newPost, nil
}

func (pr *PostDBRepo) DeletePost(ctx context.Context, postID string) error {
	bsonID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	res, err := pr.postsColl.DeleteOne(ctx, bson.M{"_id": bsonID})
	if err == nil && res.DeletedCount == 0 {
		err = fmt.Errorf("failed ro delete post %s: post not found", postID)
	}

	return err
}

func (pr *PostDBRepo) Post(ctx context.Context, postID string) (*Post, error) {
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

func (pr *PostDBRepo) PostsByCategory(ctx context.Context, category string) (*[]Post, error) {
	return pr.listByKeyValue(ctx, "category", category)
}

func (pr *PostDBRepo) CreateComment(
	ctx context.Context,
	postID string,
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

	bsonID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonID}
	update := bson.M{"$push": bson.M{"comments": comment}}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = pr.postsColl.FindOneAndUpdate(ctx, filter, update, opt).Decode(updatedPost)

	return updatedPost, err
}

func (pr *PostDBRepo) DeleteComment(ctx context.Context, postID string, commentID string, username string) (*Post, error) {
	commentBSONID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return nil, err
	}

	updatedPost := &Post{}

	bsonID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonID}
	update := bson.M{"$pull": bson.M{"comments": bson.M{"_id": commentBSONID}}}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = pr.postsColl.FindOneAndUpdate(ctx, filter, update, opt).Decode(updatedPost)

	return updatedPost, err
}

func (pr *PostDBRepo) Vote(ctx context.Context, postID string, username string, voteVal int) (*Post, error) {
	bsonID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bsonID}

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

func (pr *PostDBRepo) PostsByUser(ctx context.Context, username string) (*[]Post, error) {
	return pr.listByKeyValue(ctx, "author.username", username)
}

func (pr *PostDBRepo) listByKeyValue(ctx context.Context, key, value string) (*[]Post, error) {
	var catPosts []Post
	cur, err := pr.postsColl.Find(ctx, bson.M{key: value})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &catPosts)

	return &catPosts, err
}
