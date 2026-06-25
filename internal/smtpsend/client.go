package smtpsend

import "context"

// Client handles SMTP message submission.
type Client struct{}

func NewClient() *Client { return &Client{} }

func (c *Client) Send(ctx context.Context, req any) error { return nil }
