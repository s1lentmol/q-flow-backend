package queue

import (
	"context"

	queuev1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/queue"
	"google.golang.org/grpc"
)

type Client struct {
	api queuev1.QueueClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		api: queuev1.NewQueueClient(conn),
	}
}

func (c *Client) List(ctx context.Context, group string) ([]*queuev1.QueueDTO, error) {
	resp, err := c.api.ListQueues(ctx, &queuev1.ListQueuesRequest{GroupCode: group})
	if err != nil {
		return nil, err
	}
	return resp.GetQueues(), nil
}

func (c *Client) Create(ctx context.Context, req *queuev1.CreateQueueRequest) (*queuev1.QueueDTO, error) {
	resp, err := c.api.CreateQueue(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetQueue(), nil
}

func (c *Client) Get(ctx context.Context, queueID int64, group string) (*queuev1.GetQueueResponse, error) {
	return c.api.GetQueue(ctx, &queuev1.GetQueueRequest{QueueId: queueID, GroupCode: group})
}

func (c *Client) Join(ctx context.Context, queueID, userID int64, fullName string, group string, slotTime string) (int32, error) {
	resp, err := c.api.JoinQueue(ctx, &queuev1.JoinQueueRequest{
		QueueId:   queueID,
		UserId:    userID,
		UserName:  fullName,
		GroupCode: group,
		SlotTime:  slotTime,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetPosition(), nil
}

func (c *Client) Leave(ctx context.Context, queueID, userID int64, group string) error {
	_, err := c.api.LeaveQueue(ctx, &queuev1.LeaveQueueRequest{
		QueueId:   queueID,
		UserId:    userID,
		GroupCode: group,
	})
	return err
}

func (c *Client) Advance(ctx context.Context, queueID, actorID int64, group string) (*queuev1.ParticipantDTO, error) {
	resp, err := c.api.AdvanceQueue(ctx, &queuev1.AdvanceQueueRequest{
		QueueId:   queueID,
		GroupCode: group,
		ActorId:   actorID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetRemoved(), nil
}

func (c *Client) Remove(ctx context.Context, queueID, userID, actorID int64, group string) error {
	_, err := c.api.RemoveParticipant(ctx, &queuev1.RemoveParticipantRequest{
		QueueId:   queueID,
		UserId:    userID,
		ActorId:   actorID,
		GroupCode: group,
	})
	return err
}

func (c *Client) Archive(ctx context.Context, queueID, actorID int64, group string) error {
	_, err := c.api.ArchiveQueue(ctx, &queuev1.ArchiveQueueRequest{
		QueueId:   queueID,
		GroupCode: group,
		ActorId:   actorID,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, queueID, actorID int64, group string) error {
	_, err := c.api.DeleteQueue(ctx, &queuev1.DeleteQueueRequest{
		QueueId:   queueID,
		GroupCode: group,
		ActorId:   actorID,
	})
	return err
}

func (c *Client) Update(ctx context.Context, queueID, actorID int64, group string, title, description string) (*queuev1.QueueDTO, error) {
	resp, err := c.api.UpdateQueue(ctx, &queuev1.UpdateQueueRequest{
		QueueId:     queueID,
		GroupCode:   group,
		ActorId:     actorID,
		Title:       title,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetQueue(), nil
}

func (c *Client) Add(ctx context.Context, queueID, userID, actorID int64, fullName string, group string, slotTime string) (int32, error) {
	resp, err := c.api.AddParticipant(ctx, &queuev1.AddParticipantRequest{
		QueueId:   queueID,
		UserId:    userID,
		ActorId:   actorID,
		GroupCode: group,
		UserName:  fullName,
		SlotTime:  slotTime,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetPosition(), nil
}
