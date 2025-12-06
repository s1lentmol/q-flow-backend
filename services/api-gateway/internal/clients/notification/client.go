package notification

import (
	"context"

	notificationv1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/notification"
	"google.golang.org/grpc"
)

type Client struct {
	api notificationv1.NotificationClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		api: notificationv1.NewNotificationClient(conn),
	}
}

func (c *Client) SetContact(ctx context.Context, userID int64, username, chatID string) error {
	_, err := c.api.SetContact(ctx, &notificationv1.SetContactRequest{
		UserId:           userID,
		TelegramUsername: username,
		ChatId:           chatID,
	})
	return err
}
