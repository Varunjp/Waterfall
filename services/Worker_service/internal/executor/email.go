package executor

import (
	"context"
	"encoding/json"
	"fmt"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailExecutor struct{}

func (e *EmailExecutor) Execute(ctx context.Context, payload []byte) error {
	var p EmailPayload
	if err := json.Unmarshal(payload,&p); err != nil {
		return err 
	}


	// simulate email send

	fmt.Println("email send to :",p.To)

	return nil 
}


func init() {
	Register("email",&EmailExecutor{})
}