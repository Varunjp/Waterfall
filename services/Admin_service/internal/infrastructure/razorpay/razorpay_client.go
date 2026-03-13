package razorpayclient

import (
	"os"

	"github.com/razorpay/razorpay-go"
)

func NewRazorpayClient() *razorpay.Client {

	key := os.Getenv("RAZORPAY_KEY_ID")
	secret := os.Getenv("RAZORPAY_KEY_SECRET")

	client := razorpay.NewClient(key,secret)

	return client
}