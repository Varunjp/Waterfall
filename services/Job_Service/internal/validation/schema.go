package validator

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)
func Validate(schema string, payload []byte) error {
	s := gojsonschema.NewStringLoader(schema)
	d := gojsonschema.NewBytesLoader(payload)

	res, err := gojsonschema.Validate(s, d)
	if err != nil {
		return err
	}

	if !res.Valid() {
		return errors.New("invalid payload schema")
	}

	return nil
}