package stripeClient

import (
	"github.com/stripe/stripe-go/v78"
)

func InitStripe(secret string) {
	stripe.Key = secret
}