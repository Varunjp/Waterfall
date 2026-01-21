package producer

import "context"

type Producer interface {
	Publish(ctx context.Context, key string, value any) error 
}
