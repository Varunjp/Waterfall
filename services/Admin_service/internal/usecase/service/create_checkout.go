package service

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/repository/interfaces"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	sub "github.com/stripe/stripe-go/v78/subscription"
)

type BillingService struct {
	repo interfaces.BillingRepository
	cfg  Config
	redis *redis.Client
}

type Config struct {
	Stripe struct {
		SuccessURL string
		CancelURL  string 
	}
}

type CreateCheckoutRequest struct {
	AppID	string 
	PlanID 	string 
}

type CreateCheckoutResponse struct {
	SessionID 	string 
	CheckoutURL string 
}

func NewBillingService(r interfaces.BillingRepository,c struct{
	Stripe struct {
		SuccessURL string
		CancelURL  string
	}
},rd *redis.Client) *BillingService {
	return &BillingService{
		repo: r,
		cfg: c,
		redis: rd,
	}
}

func (s *BillingService) CreateChecoutSession(ctx context.Context,req CreateCheckoutRequest)(*CreateCheckoutResponse,error) {
	if req.AppID == "" {
		return nil,errors.New("app_id is required")
	}

	if req.PlanID == "" {
		return nil,errors.New("plan_id is required")
	}

	plan,err := s.repo.GetPlanByID(ctx,req.PlanID)
	if err != nil {
		return nil,err 
	}

	if plan == nil {
		return nil,errors.New("plan not found")
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),

		SuccessURL: stripe.String(s.cfg.Stripe.SuccessURL),
		CancelURL: stripe.String(s.cfg.Stripe.CancelURL),

		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price: stripe.String(plan.StripeID),
				Quantity: stripe.Int64(1),
			},
		},

		Metadata: map[string]string{
			"app_id": req.AppID,
			"plan_id": req.PlanID,
		},

		AllowPromotionCodes: stripe.Bool(true),
	}

	sess,err := session.New(params)
	if err != nil {
		return nil,err 
	}

	resp := &CreateCheckoutResponse{
		SessionID: sess.ID,
		CheckoutURL: sess.URL,
	}

	return resp,nil 
}

func (s *BillingService) ActivateSubscription(ctx context.Context,appID,planID,stripeSubID string) error {
	stripeSub,err := sub.Get(stripeSubID,nil)
	if err != nil {
		return err 
	}

	start := time.Unix(stripeSub.CurrentPeriodStart,0)
	end := time.Unix(stripeSub.CurrentPeriodEnd,0)

	subscription := entities.Subscription{
		AppID: appID,
		PlanID: planID,
		StripeSubscriptionID: stripeSubID,
		Status: string(stripeSub.Status),
		CurrentPeriodStart: start,
		CurrentPeriodEnd: end,
		CreatedAt: time.Now(),
	}

	err = s.repo.CreateSubscription(ctx, &subscription)

	if err != nil {
		return err
	}

	err = s.repo.UpdateAppPlan(ctx, appID, planID)
	if err != nil {
		return err
	}

	return nil
}

func (s *BillingService) CancelSubscription(
	ctx context.Context,
	stripeSubID string,
) error {

	params := &stripe.SubscriptionCancelParams{}
	_, err := sub.Cancel(stripeSubID, params)
	if err != nil {
		return err
	}

	err = s.repo.UpdateSubscriptionStatus(
		ctx,
		stripeSubID,
		"canceled",
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *BillingService) HandlePaymentSuccess(
	ctx context.Context,
	stripeSubID string,
	appID string, 
) error {

	stripeSub, err := sub.Get(stripeSubID, nil)
	if err != nil {
		return err
	}

	start := time.Unix(stripeSub.CurrentPeriodStart, 0)
	end := time.Unix(stripeSub.CurrentPeriodEnd, 0)

	err = s.repo.UpdateSubscriptionPeriod(
		ctx,
		stripeSubID,
		start,
		end,
	)

	if err != nil {
		return err
	}

	err = s.repo.ResetMonthlyUsage(
		ctx,
		appID,
	)

	if err != nil {
		return err
	}

	key := fmt.Sprintf("usage:%s:%s",appID,time.Now().Format("2006-01"))

	err = s.redis.Del(key).Err()

	if err != nil && err != redis.Nil{
		return err 
	}

	planKey := fmt.Sprintf("plan:%s",appID)

	err = s.redis.Del(planKey).Err()

	if err != nil && err != redis.Nil{
		return err 
	}

	return nil
}

func (s *BillingService) HandlePaymentFailure(
	ctx context.Context,
	stripeSubID string,
) error {

	err := s.repo.UpdateSubscriptionStatus(
		ctx,
		stripeSubID,
		"past_due",
	)

	if err != nil {
		return err
	}

	err = s.repo.BlockAppBilling(
		ctx,
		stripeSubID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *BillingService)GetSubscription(ctx context.Context,appID string) (*entities.Subscription,error) {

	substription,err := s.repo.GetSubscription(ctx,appID)

	if err != nil {
		return nil,err 
	}

	return substription,nil 
}