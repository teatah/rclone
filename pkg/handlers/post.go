package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	postpkg "gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/post"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/responses"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"

	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"

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

	posts, err := ph.PostRepo.List(r.Context())
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
	user, err := ph.UserRepo.GetUserById(r.Context(), sess.UserId)
	if err != nil {
		rc.HandleError(err)
		return
	}
	newPost, err := ph.PostRepo.Create(r.Context(), user, postRequest)
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

	post, err := ph.PostRepo.Get(r.Context(), postID)
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

	posts, err := ph.PostRepo.GetByCategory(r.Context(), category)
	if err != nil {
		rc.HandleError(err)
		return
	}

	rc.WriteRawDataToBody(posts)
}

func (ph *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postId := vars["postID"]

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

	user, err := ph.UserRepo.GetUserById(r.Context(), sess.UserId)
	if err != nil {
		rc.HandleError(err)
		return
	}

	modifiedPost, err := ph.PostRepo.CreateComment(r.Context(), postId, commentRequest.Comment, user)
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
	postId := vars["postID"]
	commentId := vars["commentID"]

	sess, err := SessionFromContext(r)
	if err != nil {
		rc.HandleError(err)
		return
	}

	username, err := ph.SessionManager.UsernameBySessionID(r.Context(), sess.ID)
	if err != nil {
		rc.HandleError(err)
		return
	}

	modifiedPost, err := ph.PostRepo.DeleteComment(r.Context(), postId, commentId, username)
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
	postId := vars["postID"]

	sess, err := SessionFromContext(rc.Request)
	if err != nil {
		return nil, err
	}

	modifiedPost, err := ph.PostRepo.Vote(ctx, postId, sess.UserId, voteVal)

	return modifiedPost, err
}

func (ph *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: ph.Logger, Writer: w, Request: r}

	vars := mux.Vars(r)
	postId := vars["postID"]

	err := ph.PostRepo.Delete(r.Context(), postId)
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
