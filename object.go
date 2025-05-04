package bogo

import (
	"bytes"
	"errors"
)

var objEncErr = errors.New("object encoder error")

func encodeObject(v any) ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteByte(TypeObject)
	switch v.(type) {
	case map[string]any:
	}
	return nil, wrapError(objEncErr, "object type not supported")

}
