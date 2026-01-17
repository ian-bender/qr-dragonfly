package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitoTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/stripe/stripe-go/v81"
)

type createCheckoutSessionRequest struct {
	Plan string `json:"plan"` // "basic" or "enterprise"
}

type checkoutSessionResponse struct {
	SessionID  string `json:"sessionId"`
	SessionURL string `json:"url"`
}

// handleCreateCheckoutSession creates a Stripe Checkout session for subscription
func (srv *Server) handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	access, _ := readCookie(r, "access_token")
	if access == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return
	}

	if srv.StripeClient == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "stripe_not_configured"})
		return
	}

	// Get user from access token
	user, err := getUserFromAccessToken(ctx, srv.Cognito, access)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return
	}

	var req createCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	req.Plan = strings.TrimSpace(strings.ToLower(req.Plan))
	if req.Plan != "basic" && req.Plan != "enterprise" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_plan"})
		return
	}

	priceID, err := srv.StripeClient.GetPriceIDForPlan(req.Plan)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_plan"})
		return
	}

	checkoutSession, err := srv.StripeClient.CreateCheckoutSession(user.Email, priceID)
	if err != nil {
		log.Printf("stripe checkout error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "checkout_failed"})
		return
	}

	writeJSON(w, http.StatusOK, checkoutSessionResponse{
		SessionID:  checkoutSession.ID,
		SessionURL: checkoutSession.URL,
	})
}

// handleCreatePortalSession creates a Stripe Customer Portal session for subscription management
func (srv *Server) handleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	access, _ := readCookie(r, "access_token")
	if access == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return
	}

	if srv.StripeClient == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "stripe_not_configured"})
		return
	}

	// Get user from access token
	user, err := getUserFromAccessToken(ctx, srv.Cognito, access)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not_authenticated"})
		return
	}

	portalSession, err := srv.StripeClient.CreateCustomerPortalSession(user.Email)
	if err != nil {
		log.Printf("stripe portal session error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "portal_session_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"url": portalSession.URL,
	})
}

// handleStripeWebhook processes Stripe webhook events
func (srv *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if srv.StripeClient == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "stripe_not_configured"})
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := srv.StripeClient.ConstructEvent(payload, signature)
	if err != nil {
		log.Printf("webhook signature verification failed: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		srv.handleCheckoutCompleted(event)
	case "customer.subscription.updated":
		srv.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		srv.handleSubscriptionDeleted(event)
	default:
		log.Printf("unhandled webhook event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func (srv *Server) handleCheckoutCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("error parsing checkout.session.completed: %v", err)
		return
	}

	if session.Mode != stripe.CheckoutSessionModeSubscription {
		return
	}

	customerEmail := session.CustomerEmail
	if customerEmail == "" && session.CustomerDetails != nil {
		customerEmail = session.CustomerDetails.Email
	}

	if customerEmail == "" {
		log.Printf("no email in checkout session: %s", session.ID)
		return
	}

	// Determine subscription tier from line items
	userType := "basic" // default
	if session.Subscription != nil {
		sub := session.Subscription
		if sub.Items != nil && len(sub.Items.Data) > 0 {
			// You can check price ID to determine tier
			// For now, we'll need metadata or price ID mapping
		}
	}

	log.Printf("checkout completed for %s, updating to %s", customerEmail, userType)
	srv.updateUserTypeByEmail(context.Background(), customerEmail, userType)
}

func (srv *Server) handleSubscriptionUpdated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("error parsing customer.subscription.updated: %v", err)
		return
	}

	if subscription.Status != stripe.SubscriptionStatusActive {
		log.Printf("subscription %s not active, status: %s", subscription.ID, subscription.Status)
		return
	}

	// Get customer email
	if subscription.Customer == nil {
		return
	}

	// Note: In production, you'd fetch customer details or store mapping
	log.Printf("subscription %s updated, customer: %v", subscription.ID, subscription.Customer)
}

func (srv *Server) handleSubscriptionDeleted(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("error parsing customer.subscription.deleted: %v", err)
		return
	}

	// Downgrade user to free tier
	log.Printf("subscription %s deleted, customer should be downgraded", subscription.ID)
}

func (srv *Server) updateUserTypeByEmail(ctx context.Context, email, userType string) {
	// List users to find by email
	listOut, err := srv.Cognito.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(srv.UserPoolID),
		Filter:     aws.String(fmt.Sprintf(`email = "%s"`, email)),
		Limit:      aws.Int32(1),
	})
	if err != nil || len(listOut.Users) == 0 {
		log.Printf("user not found for email %s: %v", email, err)
		return
	}

	username := aws.ToString(listOut.Users[0].Username)

	_, err = srv.Cognito.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
		UserPoolId: aws.String(srv.UserPoolID),
		Username:   aws.String(username),
		UserAttributes: []cognitoTypes.AttributeType{
			{Name: aws.String(cognitoUserTypeAttr), Value: aws.String(userType)},
		},
	})
	if err != nil {
		log.Printf("failed to update user type for %s: %v", email, err)
		return
	}

	log.Printf("updated user %s to type %s", email, userType)
}
