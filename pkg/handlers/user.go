package handlers

import (
	"net/http"

	"github.com/teatah/rclone/pkg/responses"
	"github.com/teatah/rclone/pkg/session"
	userpkg "github.com/teatah/rclone/pkg/user"
	"go.uber.org/zap"
)

type UserHandler struct {
	SessionManager *session.DBSessionManager
	Logger         *zap.SugaredLogger
	UserRepo       userpkg.UserRepo
}

func (uh *UserHandler) GetLogger() *zap.SugaredLogger {
	return uh.Logger
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: uh.Logger, Writer: w, Request: r}

	userRequest := &userpkg.UserRequest{}
	err := responses.ReadBody(r, userRequest)
	if err != nil {
		rc.HandleError(err)
		return
	}

	ctx := r.Context()
	user, err := uh.UserRepo.Register(ctx, userRequest)
	if err != nil {
		if err == userpkg.ErrUserAlreadyExists {
			respErr := responses.NewResponseError("body", "username", "username", err.Error())
			rc.JSONError(http.StatusUnprocessableEntity, respErr)
			return
		}
		rc.HandleError(err)
		return
	}

	sess, err := uh.SessionManager.Create(ctx, user)
	if err != nil {
		rc.HandleError(err)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	bodyResponse := responses.TokenResponse{Token: sess.ID}
	rc.WriteRawDataToBody(bodyResponse)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	rc := &responses.ResponseContext{Logger: uh.Logger, Writer: w, Request: r}

	userRequest := &userpkg.UserRequest{}
	err := responses.ReadBody(r, userRequest)
	if err != nil {
		rc.HandleError(err)
		return
	}

	ctx := r.Context()
	user, err := uh.UserRepo.Login(ctx, userRequest)
	if err != nil {
		rc.LogError(err)

		w.WriteHeader(http.StatusUnauthorized)
		rc.WriteRawDataToBody(responses.Message{Message: err.Error()})

		return
	}

	sess, err := uh.SessionManager.Create(ctx, user)
	if err != nil {
		rc.HandleError(err)
		return
	}

	bodyResponse := responses.TokenResponse{Token: sess.ID}
	rc.WriteRawDataToBody(bodyResponse)
}
