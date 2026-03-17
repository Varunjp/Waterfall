package routes

import (
	"admin_service/internal/middleware"
	controller "admin_service/internal/transport/rest"

	"github.com/go-chi/chi/v5"
)

func RgisterBillingRoutes(r chi.Router,billingController *controller.BillingController) {

	r.Route("/billing",func(r chi.Router){

		r.Post("/webhook",billingController.StripeWebhook)

		r.Group(func(r chi.Router){
			r.Use(middleware.AuthMiddleware)

			r.Post("/checkout",billingController.CreateCheckout)
			r.Post("/cancel",billingController.CancelSubscription)
			r.Get("/subscription",billingController.GetSubscription)
		})
	})
}