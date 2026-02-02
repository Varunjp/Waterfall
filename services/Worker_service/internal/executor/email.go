package executor

import (
	"context"
	"log"
)

func SendEmail(ctx context.Context, payload []byte) error {
	// stimulate send
	log.Println("sending email:", string(payload))
	return nil
}

func init() {
	Register("email", SendEmail)
}
