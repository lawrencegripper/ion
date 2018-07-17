package logger

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

type RedactorFunc func([]byte) []byte

type Redactor struct {
	backend logrus.Formatter

	Redactor RedactorFunc
}

func (redactor Redactor) Format(entry *logrus.Entry) ([]byte, error) {
	serialized, err := redactor.backend.Format(entry)

	if err == nil {
		serialized = redactor.Redactor(serialized)
	}

	return serialized, err
}

func (redactor *Redactor) init() {
	redactor.Redactor = func(in []byte) []byte { return in }
}

func NewJsonRedactor() Redactor {
	redactor := Redactor{}
	redactor.backend = new(logrus.JSONFormatter)
	return redactor
}

func NewJsonSecretRedactor(fun RedactorFunc) Redactor {
	redactor := Redactor{}
	redactor.backend = new(logrus.JSONFormatter)
	redactor.Redactor = fun
	return redactor
}

func NewTextRedactor() Redactor {
	redactor := Redactor{}
	redactor.backend = new(logrus.TextFormatter)
	return redactor
}

func NewTextSecretRedactor(fun RedactorFunc) Redactor {
	redactor := Redactor{}
	redactor.backend = new(logrus.TextFormatter)
	redactor.Redactor = fun
	return redactor
}

func NewSecretRedact(secrets [][]byte, redacted []byte) RedactorFunc {
	return func(serialized []byte) []byte {
		out := serialized
		for _, s := range secrets {
			out = bytes.Replace(out, s, redacted, -1)
		}
		return out
	}
}
