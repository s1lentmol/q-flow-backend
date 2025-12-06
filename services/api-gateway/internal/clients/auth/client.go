package auth

import (
	"context"

	authv1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/auth"
	"google.golang.org/grpc"
)

type Client struct {
	api authv1.AuthClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		api: authv1.NewAuthClient(conn),
	}
}

func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
	resp, err := c.api.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetUserId(), nil
}

func (c *Client) Login(ctx context.Context, email, password string, appID int32) (string, error) {
	resp, err := c.api.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	if err != nil {
		return "", err
	}
	return resp.GetToken(), nil
}
