package producer

import "context"

type Producer interface {
	Publish(ctx context.Context, value any) error 
}