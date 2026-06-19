package controller

import (
	"admin_service/internal/config"
	"admin_service/internal/middleware"
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
	cfg     *config.Config
}

func NewBillingController(s service.BillingService, cfg *config.Config) *BillingController {
	return &BillingController{
		service: s,
		cfg:     cfg,
	}
}

type CheckoutRequest struct {
	PlanID string `json:"plan_id"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func (c *BillingController) CreateCheckout(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	var req CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	appID := middleware.GetAppID(ctx)

	if appID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	url, err := c.service.CreateChecoutSession(
		ctx,
		service.CreateCheckoutRequest{
			AppID:  appID,
			PlanID: req.PlanID,
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CheckoutResponse{
		CheckoutURL: url.CheckoutURL,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("response error :",err)
	}
}

func (c *BillingController) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if c.cfg.Stripe.WebhookSecret == "" {
		log.Println("stripe webhook config error: STRIPE_WEBHOOK_SECRET is not set")
		http.Error(w, "webhook not configured", http.StatusInternalServerError)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("read error:", err)
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		log.Println("stripe webhook signature error: missing Stripe-Signature header")
		http.Error(w, "missing signature", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		payload,
		sigHeader,
		c.cfg.Stripe.WebhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)

	if err != nil {
		log.Printf("stripe webhook signature error: %v (payload_bytes=%d)", err, len(payload))
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Println("checkout.session.completed unmarshal error:", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		appID := session.Metadata["app_id"]
		planID := session.Metadata["plan_id"]

		if appID == "" || planID == "" {
			log.Println("checkout.session.completed missing app metadata, ignoring session:", session.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}
		if subscriptionID == "" {
			log.Println("checkout.session.completed missing subscription:", session.ID)
			http.Error(w, "missing subscription", http.StatusBadRequest)
			return
		}

		err = c.service.ActivateSubscription(
			context.Background(),
			appID,
			planID,
			subscriptionID,
		)

		if err != nil {
			log.Println("activate subscription error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = c.service.HandlePaymentSuccess(
			context.Background(),
			subscriptionID,
			appID,
		)

		if err != nil {
			log.Println("payment success handler error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "invoice.payment_succeeded":

		var raw map[string]interface{}
		
		if err := json.Unmarshal(event.Data.Raw, &raw); err != nil {
			log.Println("unmarshal error:", err)
			break
		}

		if raw["billin_reason"] == "subscription_create" {
			log.Println("skipping initial invoice, already handled by checkout.session.completed")
        	break
		}

		subscriptionID := ""
		if parent, ok := raw["parent"].(map[string]interface{}); ok {
			if subDetails, ok := parent["subscription_details"].(map[string]interface{}); ok {
				subscriptionID, _ = subDetails["subscription"].(string)
			}
		}


		invoiceNumber,_ := raw["number"].(string)
		amountPaid,_:= raw["amount_paid"].(float64)

		if subscriptionID == "" {
			log.Println("could not extract subscription ID from invoice")
			break
		}

		err := c.service.SendInvoicePdf(context.Background(),subscriptionID,invoiceNumber,amountPaid)

		if err != nil {
			log.Println("failed to send invoice pdf :",err)
			break 
		}

	case "invoice.payment_failed":

		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			log.Println("invoice.payment_failed unmarshal error:", err)
			break
		}
		if invoice.Subscription == nil {
			log.Println("subscription missing in failed invoice")
			break
		}

		err := c.service.HandlePaymentFailure(
			context.Background(),
			invoice.Subscription.ID,
		)

		if err != nil {
			log.Println("payment failure handler error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Println("customer.subscription.deleted unmarshal error:", err)
			break
		}

		err := c.service.CancelSubscription(
			context.Background(),
			sub.ID,
		)

		if err != nil {
			log.Println("cancel subscription handler error:", err)
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

	appID := middleware.GetAppID(ctx)

	subscription, err := c.service.GetSubscription(ctx, appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subscription); err != nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return 
	}
}
