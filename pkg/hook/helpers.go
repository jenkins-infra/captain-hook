package hook

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

func writeResult(w io.Writer, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		logrus.Debugf("failed to write message: %s, err: %s", message, err)
	}
}

func responseHTTPError(w http.ResponseWriter, statusCode int, message string, args ...interface{}) {
	response := fmt.Sprintf(message, args...)

	logrus.WithFields(logrus.Fields{
		"response":    response,
		"status-code": statusCode,
	}).Info(response)

	http.Error(w, response, statusCode)
}
