package responses

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type ResponseError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value,omitempty"`
	Msg      string `json:"msg"`
}

func (re *ResponseError) Error() string {
	return fmt.Sprintf(
		"Location: %s, param: %s, value: %s, msg: %s",
		re.Location, re.Param, re.Value, re.Msg,
	)
}

func NewResponseError(loc, param, val, msg string) *ResponseError {
	return &ResponseError{
		Location: loc,
		Param:    param,
		Value:    val,
		Msg:      msg,
	}
}

type ResponseErrors struct {
	Errors []*ResponseError `json:"errors"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type Message struct {
	Message string `json:"message"`
}

type ResponseContext struct {
	Logger  *zap.SugaredLogger
	Writer  http.ResponseWriter
	Request *http.Request
}

func (rc *ResponseContext) WriteRawDataToBody(body any) {
	rawBody, err := ToJSON(body)
	if err != nil {
		rc.HandleError(err)
		return
	}

	_, err = rc.Writer.Write(rawBody)
	if err != nil {
		rc.HandleError(err)
	}
}

func (rc *ResponseContext) HandleError(err error) {
	rc.LogError(err)

	respErr := &ResponseError{
		Location: "body",
		Param:    "error",
		Msg:      "internal error",
	}

	rc.JSONError(http.StatusInternalServerError, respErr)
}

func (rc *ResponseContext) LogError(err error) {
	errString := err.Error()

	rc.Logger.Errorw(errString,
		"method", rc.Request.Method,
		"remote_addr", rc.Request.RemoteAddr,
		"url", rc.Request.URL.Path,
	)
}

func (rc *ResponseContext) JSONError(status int, respErr ...*ResponseError) {
	errs := ResponseErrors{
		Errors: respErr,
	}

	marshalledErrs, err := ToJSON(errs)

	if err != nil {
		rc.HandleError(err)
	}

	http.Error(rc.Writer, string(marshalledErrs), status)
}

func ReadBody(r *http.Request, receiver any) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, receiver)

	return err
}

func ToJSON(body any) ([]byte, error) {
	resp, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return resp, err
}
