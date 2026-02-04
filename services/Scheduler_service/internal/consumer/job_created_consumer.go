package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type JobCreatedConsumer struct {
	reader *kafka.Reader
}


func NewJobCreatedConsumer(brokers,topic,groupID string) *JobCreatedConsumer {
	return &JobCreatedConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{brokers},
			Topic: topic,
			GroupID: groupID,
		}),
	}
}

func (c *JobCreatedConsumer) Read(ctx context.Context) ([]byte,error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil,err 
	}
	return msg.Value,nil 
}