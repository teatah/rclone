package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	postpkg "github.com/teatah/rclone/pkg/post"
	"github.com/teatah/rclone/pkg/responses"
	"github.com/teatah/rclone/pkg/session"

	"github.com/teatah/rclone/pkg/user"

	"go.uber.org/zap"
)

type PostHandler struct {
	SessionManager *session.DBSessionManager
	Logger         *zap.SugaredLogger
	PostRepo       postpkg.PostRepo
	UserRepo       user.UserRepo
}

func (ph *PostHandler) Posts(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	posts, err := ph.PostRepo.AllPosts(r.Context())
	if err != nil {
		rc.HandleError(err)
		return
	}
	rc.WriteRawDataToBody(posts)
}

func (ph *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	postRequest := &postpkg.PostRequest{}
	err := responses.ReadBody(r, postRequest)
	if err != nil {
		rc.HandleError(err)
		return
	}

	sess, err := SessionFromContext(r)
	if err != nil {
		rc.HandleError(err)
		return
	}

	ctx := r.Context()
	user, err := ph.UserRepo.UserByID(ctx, sess.UserID)
	if err != nil {
		rc.HandleError(err)
		return
	}
	newPost, err := ph.PostRepo.CreatePost(ctx, user, postRequest)
	if err != nil {
		rc.HandleError(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	rc.WriteRawDataToBody(newPost)
}

func (ph *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postID := vars["postID"]

	post, err := ph.PostRepo.Post(r.Context(), postID)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(post)
}

func (ph *PostHandler) PostsByCategory(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	category := vars["category"]

	posts, err := ph.PostRepo.PostsByCategory(r.Context(), category)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(posts)
}

func (ph *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postID := vars["postID"]

	commentRequest := &postpkg.CommentRequest{}
	err := responses.ReadBody(r, commentRequest)
	if err != nil {
		rc.HandleError(err)
		return
	}

	if len(commentRequest.Comment) == 0 {
		err = &responses.ResponseError{
			Location: "body",
			Param:    "comment",
			Msg:      "is required",
		}
		respErr, _ := err.(*responses.ResponseError)
		rc.JSONError(http.StatusUnprocessableEntity, respErr)
		return
	}

	sess, err := SessionFromContext(r)
	if err != nil {
		rc.HandleError(err)
		return
	}

	ctx := r.Context()
	user, err := ph.UserRepo.UserByID(ctx, sess.UserID)
	if err != nil {
		rc.HandleError(err)
		return
	}

	modifiedPost, err := ph.PostRepo.CreateComment(ctx, postID, commentRequest.Comment, user)
	if err != nil {
		rc.HandleError(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	rc.WriteRawDataToBody(modifiedPost)
}

func (ph *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postID := vars["postID"]
	commentID := vars["commentID"]

	sess, err := SessionFromContext(r)
	if err != nil {
		rc.HandleError(err)
		return
	}

	ctx := r.Context()
	username, err := ph.SessionManager.UsernameBySessionID(ctx, sess.ID)
	if err != nil {
		rc.HandleError(err)
		return
	}

	modifiedPost, err := ph.PostRepo.DeleteComment(ctx, postID, commentID, username)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(modifiedPost)
}

func (ph *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	modifiedPost, err := ph.vote(r.Context(), rc, postpkg.Upvote)
	if err != nil {
		rc.HandleError(err)
		return
	}
	rc.WriteRawDataToBody(modifiedPost)
}

func (ph *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	modifiedPost, err := ph.vote(r.Context(), rc, postpkg.Downnvote)
	if err != nil {
		rc.HandleError(err)
		return
	}
	rc.WriteRawDataToBody(modifiedPost)
}

func (ph *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	modifiedPost, err := ph.vote(r.Context(), rc, postpkg.Unvote)
	if err != nil {
		rc.HandleError(err)
		return
	}
	rc.WriteRawDataToBody(modifiedPost)
}

func (ph *PostHandler) vote(ctx context.Context, rc *responses.ResponseContext, voteVal int) (*postpkg.Post, error) {
	vars := mux.Vars(rc.Request)
	postID := vars["postID"]

	sess, err := SessionFromContext(rc.Request)
	if err != nil {
		return nil, err
	}

	modifiedPost, err := ph.PostRepo.Vote(ctx, postID, sess.UserID, voteVal)

	return modifiedPost, err
}

func (ph *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postID := vars["postID"]

	err := ph.PostRepo.DeletePost(r.Context(), postID)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(responses.Message{Message: "success"})
}

func (ph *PostHandler) PostsByUser(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	username := vars["username"]

	posts, err := ph.PostRepo.PostsByUser(r.Context(), username)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(posts)
}

func SessionFromContext(r *http.Request) (*session.Session, error) {
	sessVal := session.SessionCtxValue("session")
	ctxSess := r.Context().Value(sessVal)
	if ctxSess == nil {
		return nil, errors.New("failed to get session from context")
	}
	sess := ctxSess.(*session.Session)

	return sess, nil
}
