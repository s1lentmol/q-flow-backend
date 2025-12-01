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
	return &Client{api: notificationv1.NewNotificationClient(conn)}
}

func (c *Client) NotifyPositionSoon(ctx context.Context, userID int64, queueTitle string, position int32) error {
	_, err := c.api.NotifyPositionSoon(ctx, &notificationv1.NotifyPositionSoonRequest{
		UserId:     userID,
		QueueTitle: queueTitle,
		Position:   position,
	})
	return err
}
