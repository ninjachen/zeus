package httperror

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"go.yym.plus/zeus/pkg/log"
)

type HttpErrCode interface {
	Error() string
	Status() int
	Code() int
	Message() string
	Prompt() string
	WithMessage(message string) HttpErrCode
	WithError(err error) HttpErrCode
	WithPrompt(message string) HttpErrCode
}

type httpErrCode struct {
	status  int
	code    int
	message string
	prompt  string
}

var (
	codes = map[int]HttpErrCode{}
)

func (self *httpErrCode) Error() string {
	return fmt.Sprintf("status:%d,code:%d,message:%s,prompt:%s", self.status, self.code, self.message, self.prompt)
}

func (self *httpErrCode) Status() int {
	return self.status
}

func (self *httpErrCode) Code() int {
	return self.code
}

func (self *httpErrCode) Message() string {
	return self.message
}

func (self *httpErrCode) Prompt() string {
	return self.prompt
}

func (self *httpErrCode) WithMessage(message string) HttpErrCode {
	return &httpErrCode{
		code:    self.code,
		message: message,
		status:  self.status,
		prompt:  self.prompt,
	}
}

func (self *httpErrCode) WithError(err error) HttpErrCode {
	log.WithError(err).WithOptions(zap.AddCallerSkip(1)).Error("occur error")
	return self
}

func (self *httpErrCode) WithPrompt(prompt string) HttpErrCode {
	return &httpErrCode{
		code:    self.code,
		message: self.message,
		status:  self.status,
		prompt:  prompt,
	}
}

func Register(status int, code int, message string, prompt string) HttpErrCode {
	if codes[code] != nil {
		panic(errors.Errorf("http err code %d has exist!", code))
	}
	codes[code] = &httpErrCode{
		code:    code,
		message: message,
		status:  status,
		prompt:  prompt,
	}
	return codes[code]
}
