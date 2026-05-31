package polar

import (
	"context"
	"fmt"

	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
	"github.com/polarsource/polar-go/models/operations"
)

type Client struct {
	sdk *polargo.Polar
}

func NewClient(accessToken string, sandbox bool) *Client {
	opts := []polargo.SDKOption{
		polargo.WithSecurity(accessToken),
	}
	if sandbox {
		opts = append(opts, polargo.WithServer("sandbox"))
	}
	return &Client{
		sdk: polargo.New(opts...),
	}
}

func (c *Client) CreateCustomerPortalSession(ctx context.Context, externalCustomerID, returnURL string) (string, error) {
	res, err := c.sdk.CustomerSessions.Create(ctx,
		operations.CreateCustomerSessionsCreateCustomerSessionCreateCustomerSessionCustomerExternalIDCreate(
			components.CustomerSessionCustomerExternalIDCreate{
				ExternalCustomerID: externalCustomerID,
				ReturnURL:          &returnURL,
			},
		),
	)
	if err != nil {
		return "", fmt.Errorf("creating Polar customer session: %w", err)
	}

	if res.CustomerSession == nil {
		return "", fmt.Errorf("Polar customer session response missing data")
	}

	return res.CustomerSession.CustomerPortalURL, nil
}

func (c *Client) CreateCheckoutSession(ctx context.Context, productID, externalCustomerID, customerEmail, successURL string) (string, error) {
	res, err := c.sdk.Checkouts.Create(ctx, components.CheckoutCreate{
		Products:           []string{productID},
		ExternalCustomerID: &externalCustomerID,
		CustomerEmail:      &customerEmail,
		SuccessURL:         &successURL,
	})
	if err != nil {
		return "", fmt.Errorf("creating Polar checkout: %w", err)
	}

	if res.Checkout == nil {
		return "", fmt.Errorf("Polar checkout response missing checkout data")
	}

	return res.Checkout.URL, nil
}
