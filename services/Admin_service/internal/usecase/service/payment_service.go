package service

import "github.com/razorpay/razorpay-go"

type PaymentService struct {
	client *razorpay.Client
}

func NewPaymentService(client *razorpay.Client) *PaymentService {
	return &PaymentService{client: client}
}

func (s *PaymentService) CreateOrder(amount int64)(string,error) {
	data := map[string]interface{}{
		"amount": amount,
		"currency": "INR",
	}

	order,err := s.client.Order.Create(data,nil)
	if err != nil {
		return "",err 
	}

	orderID := order["id"].(string)
	return orderID,nil 
}