package testutil

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func NopLogger() logrus.FieldLogger {
	l := logrus.New()
	l.Out = ioutil.Discard

	return l
}
