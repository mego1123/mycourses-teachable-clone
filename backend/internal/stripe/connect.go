// Package stripe — Stripe Connect service for marketplace payouts.
// Handles creator onboarding, course checkout with destination charges,
// and payout management via Stripe Connect Express accounts.
package stripe

import (
        "context"
        "fmt"

        "github.com/google/uuid"
        stripe "github.com/stripe/stripe-go/v82"
        account "github.com/stripe/stripe-go/v82/account"
        accountlink "github.com/stripe/stripe-go/v82/accountlink"
        checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
        transfer "github.com/stripe/stripe-go/v82/transfer"
        transferreversal "github.com/stripe/stripe-go/v82/transferreversal"
        webhook "github.com/stripe/stripe-go/v82/webhook"

        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
)

// ConnectService handles Stripe Connect operations for the course marketplace.
type ConnectService struct {
        secretKey       string
        webhookSecret   string
        connectClientID string
        db              *db.DB
        frontendURL     string
}

// NewConnectService creates a new Stripe Connect service.
func NewConnectService(secretKey, webhookSecret, connectClientID string, database *db.DB, frontendURL string) *ConnectService {
        stripe.Key = secretKey
        return &ConnectService{
                secretKey:       secretKey,
                webhookSecret:   webhookSecret,
                connectClientID: connectClientID,
                db:              database,
                frontendURL:     frontendURL,
        }
}

// OnboardCreator creates a Stripe Express account and returns an onboarding URL.
func (s *ConnectService) OnboardCreator(ctx context.Context, tenantID string, returnURL, refreshURL string) (string, error) {
        tenantUUID, err := uuid.Parse(tenantID)
        if err != nil {
                return "", fmt.Errorf("invalid tenant ID: %w", err)
        }

        tenant, err := s.db.Queries.GetTenantByID(ctx, tenantUUID)
        if err != nil {
                return "", fmt.Errorf("tenant not found: %w", err)
        }

        var accountID string
        if tenant.StripeConnectAccountID != nil && *tenant.StripeConnectAccountID != "" {
                accountID = *tenant.StripeConnectAccountID
        } else {
                acct, err := account.New(&stripe.AccountParams{
                        Type:    stripe.String(string(stripe.AccountTypeExpress)),
                        Country: stripe.String("US"),
                        Metadata: map[string]string{
                                "tenant_id": tenantID,
                        },
                })
                if err != nil {
                        return "", fmt.Errorf("failed to create Stripe account: %w", err)
                }
                accountID = acct.ID

                if err := s.db.Queries.UpdateTenantStripeConnect(ctx, gen.UpdateTenantStripeConnectParams{
                        ID:                     tenant.ID,
                        StripeConnectAccountID: &accountID,
                        StripeConnectStatus:    "pending",
                }); err != nil {
                        return "", fmt.Errorf("failed to save Connect account ID: %w", err)
                }
        }

        link, err := accountlink.New(&stripe.AccountLinkParams{
                Account:    stripe.String(accountID),
                RefreshURL: stripe.String(refreshURL),
                ReturnURL:  stripe.String(returnURL),
                Type:       stripe.String("account_onboarding"),
        })
        if err != nil {
                return "", fmt.Errorf("failed to create account link: %w", err)
        }

        return link.URL, nil
}

// CreateCourseCheckout creates a Stripe Checkout Session for a course purchase
// with destination charges: platform takes commission, rest goes to creator.
func (s *ConnectService) CreateCourseCheckout(
        ctx context.Context,
        course *gen.Course,
        tenant *gen.Tenant,
        learnerEmail string,
        learnerUserID string,
        successURL, cancelURL string,
) (*stripe.CheckoutSession, error) {
        if tenant.StripeConnectAccountID == nil || *tenant.StripeConnectAccountID == "" {
                return nil, fmt.Errorf("creator has not completed Stripe Connect onboarding")
        }

        commissionBps := int64(tenant.CommissionRateBps)
        finalPrice := course.PriceCents
        applicationFee := finalPrice * commissionBps / 10000

        params := &stripe.CheckoutSessionParams{
                Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
                LineItems: []*stripe.CheckoutSessionLineItemParams{
                        {
                                Quantity: stripe.Int64(1),
                                PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                                        Currency: stripe.String(course.Currency),
                                        ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                                                Name: stripe.String(course.Title),
                                        },
                                        UnitAmount: stripe.Int64(finalPrice),
                                },
                        },
                },
                PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
                        ApplicationFeeAmount: stripe.Int64(applicationFee),
                        TransferData: &stripe.CheckoutSessionPaymentIntentDataTransferDataParams{
                                Destination: stripe.String(*tenant.StripeConnectAccountID),
                        },
                },
                SuccessURL: stripe.String(successURL),
                CancelURL:  stripe.String(cancelURL),
                Metadata: map[string]string{
                        "course_id":       course.ID.String(),
                        "tenant_id":       tenant.ID.String(),
                        "learner_user_id": learnerUserID,
                        "type":            "course_purchase",
                },
        }

        if learnerEmail != "" {
                params.CustomerEmail = stripe.String(learnerEmail)
        }

        session, err := checkoutsession.New(params)
        if err != nil {
                return nil, fmt.Errorf("failed to create checkout session: %w", err)
        }

        return session, nil
}

// InitiatePayout creates a Stripe Transfer to the creator's connected account.
func (s *ConnectService) InitiatePayout(ctx context.Context, tenantID string, amountCents int64, currency string) (string, error) {
        tenantUUID, err := uuid.Parse(tenantID)
        if err != nil {
                return "", fmt.Errorf("invalid tenant ID: %w", err)
        }

        tenant, err := s.db.Queries.GetTenantByID(ctx, tenantUUID)
        if err != nil {
                return "", fmt.Errorf("tenant not found: %w", err)
        }

        if tenant.StripeConnectAccountID == nil || *tenant.StripeConnectAccountID == "" {
                return "", fmt.Errorf("creator has not completed Stripe Connect onboarding")
        }

        tr, err := transfer.New(&stripe.TransferParams{
                Amount:      stripe.Int64(amountCents),
                Currency:    stripe.String(currency),
                Destination: stripe.String(*tenant.StripeConnectAccountID),
                Metadata: map[string]string{
                        "tenant_id": tenantID,
                        "type":      "creator_payout",
                },
        })
        if err != nil {
                return "", fmt.Errorf("failed to create transfer: %w", err)
        }

        return tr.ID, nil
}

// ReverseTransfer reverses a transfer (used when refunding a course purchase).
func (s *ConnectService) ReverseTransfer(transferID string, amountCents int64) error {
        _, err := transferreversal.New(&stripe.TransferReversalParams{
                ID:     stripe.String(transferID),
                Amount: stripe.Int64(amountCents),
        })
        return err
}

// GetAccountStatus retrieves the Connect account status from Stripe.
func (s *ConnectService) GetAccountStatus(accountID string) (chargesEnabled bool, payoutsEnabled bool, detailsSubmitted bool, err error) {
        acct, err := account.GetByID(accountID, &stripe.AccountParams{})
        if err != nil {
                return false, false, false, err
        }
        return acct.ChargesEnabled, acct.PayoutsEnabled, acct.DetailsSubmitted, nil
}

// VerifyWebhookSignature verifies the Stripe webhook signature and returns the event.
func (s *ConnectService) VerifyWebhookSignature(payload []byte, signature string) (stripe.Event, error) {
        return webhook.ConstructEvent(payload, signature, s.webhookSecret)
}

// WebhookSecret returns the webhook secret (used by handlers).
func (s *ConnectService) WebhookSecret() string { return s.webhookSecret }
