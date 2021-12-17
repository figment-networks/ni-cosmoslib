package util

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

func EncodeToB64(data []byte, fieldName string) (string, error) {

	// This field can contain null bytes. This data will end up getting written to postgres
	// as a JSONB type. Postgres does not support null bytes in JSONB data types. In order to write this
	// to the database while preserving the original value, encode it as base64.

	buff := bytes.NewBuffer([]byte{})
	b64Enc := base64.NewEncoder(base64.StdEncoding, buff)
	defer b64Enc.Close()

	if _, err := b64Enc.Write(data); err != nil {
		return "", fmt.Errorf("could not encode %s: %w", fieldName, err)
	}

	return buff.String(), nil
}
