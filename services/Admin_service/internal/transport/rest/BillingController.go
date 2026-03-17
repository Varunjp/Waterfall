package controller

import (
	"admin_service/internal/config"
	"admin_service/internal/usecase/service"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

type BillingController struct {
	service service.BillingService
	cfg  *config.Config
}

func NewBillingController(s service.BillingService,cfg *config.Config)*BillingController{
	return &BillingController{
		service: s,
		cfg: cfg,
	}
}

type CheckoutRequest struct {
	PlanID  string `json:"plan_id"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func (c *BillingController) CreateCheckout(w http.ResponseWriter,r *http.Request) {

	ctx := r.Context()

	var req CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"invalid request body",http.StatusBadRequest)
		return 
	}

	appID := r.Header.Get("X-App-ID")

	url,err := c.service.CreateChecoutSession(
		ctx,
		service.CreateCheckoutRequest{
			AppID: appID,
			PlanID: req.PlanID,
		},
	)

	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return 
	}

	resp := CheckoutResponse{
		CheckoutURL: url.CheckoutURL,
	}
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c *BillingController) StripeWebhook(w http.ResponseWriter,r *http.Request) {
	
	payload,err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("read error:", err)
		http.Error(w,"unable to read body",http.StatusBadRequest)
		return 
	}

	sigHeader := r.Header.Get("Stripe-Signature")

	event,err := webhook.ConstructEventWithOptions(
		payload,
		sigHeader,
		c.cfg.Stripe.WebhookSecret,
		webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
		},
	)

	if err != nil {
		log.Println("signature error:", err)
		http.Error(w,"invalid signature",http.StatusBadRequest)
		return 
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw,&session)
		if err != nil {
			http.Error(w,"invalid payload",http.StatusBadRequest)
			return 
		}

		appID := session.Metadata["app_id"]
		planID := session.Metadata["plan_id"]

		if appID == "" || planID == "" {
			log.Println("missing metadata", session.ID)
			http.Error(w, "missing metadata", http.StatusBadRequest)
			return
		}

		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}

		err = c.service.ActivateSubscription(
			context.Background(),
			appID,
			planID,
			subscriptionID,
		)

		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return 
		}

		err = c.service.HandlePaymentSuccess(
			context.Background(),
			subscriptionID,
			appID,
		)

		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw,&invoice);err != nil {
			log.Println("unmarshal error:",err)
			break 
		}

		if invoice.Subscription == nil {
			log.Println("subscription missing in invoice")
			break 
		}

		// subscriptionID := invoice.Subscription.ID

		// err := c.service.HandlePaymentSuccess(
		// 	context.Background(),
		// 	subscriptionID,
		// 	"",
		// )

		// if err != nil {
		// 	log.Println("service error:", err)
		// }
	
	case "invoice.payment_failed":

		var invoice stripe.Invoice
		json.Unmarshal(event.Data.Raw,&invoice)

		err := c.service.HandlePaymentFailure(
			context.Background(),
			invoice.Subscription.ID,
		)

		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return 
		}
	
	case "customer.subscription.deleted":
		var sub stripe.Subscription
		json.Unmarshal(event.Data.Raw, &sub)

		err := c.service.CancelSubscription(
			context.Background(),
			sub.ID,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default: 
	}

	w.WriteHeader(http.StatusOK)
}

type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id"`
}

func (c *BillingController) CancelSubscription(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx := r.Context()

	var req CancelSubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err = c.service.CancelSubscription(ctx, req.SubscriptionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *BillingController) GetSubscription(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx := r.Context()

	appID := r.Header.Get("X-App-ID")

	subscription, err := c.service.GetSubscription(ctx, appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}