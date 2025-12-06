package notification

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/s1lentmol/q-flow-backend/services/notification/internal/domain"
)

type ContactStorage interface {
	UpsertContact(ctx context.Context, contact domain.Contact) error
	GetContact(ctx context.Context, userID int64) (domain.Contact, error)
	CreateLinkToken(ctx context.Context, token domain.LinkToken) error
	ConsumeLinkToken(ctx context.Context, token string) (domain.LinkToken, error)
}

type Service struct {
	log           *slog.Logger
	storage       ContactStorage
	telegramToken string
	httpClient    *http.Client
	botName       string
}

func New(log *slog.Logger, storage ContactStorage, telegramToken string, botName string) *Service {
	return &Service{
		log:           log,
		storage:       storage,
		telegramToken: telegramToken,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		botName:       botName,
	}
}

func (s *Service) SetContact(ctx context.Context, userID int64, username, chatID string) error {
	contact := domain.Contact{
		UserID:   userID,
		Username: strings.TrimPrefix(username, "@"),
		ChatID:   chatID,
	}
	return s.storage.UpsertContact(ctx, contact)
}

func (s *Service) NotifyPositionSoon(ctx context.Context, userID int64, queueTitle string, position int32) error {
	contact, err := s.storage.GetContact(ctx, userID)
	if err != nil {
		s.log.Warn("contact not found, skip notification", slog.Int64("user_id", userID), slog.Any("err", err))
		return nil
	}

	if s.telegramToken == "" {
		s.log.Info("telegram token not set, logging notification",
			slog.Int64("user_id", userID),
			slog.String("queue", queueTitle),
			slog.Int("position", int(position)),
		)
		return nil
	}

	chat := contact.ChatID
	if chat == "" {
		if contact.Username == "" {
			s.log.Warn("no chat or username, skip telegram notification", slog.Int64("user_id", userID))
			return nil
		}
		chat = "@" + contact.Username
	}

	text := fmt.Sprintf("Очередь по \"%s\": скоро ваша очередь. Текущее место: %d.", queueTitle, position)
	return s.sendTelegramMessage(ctx, chat, text)
}

func (s *Service) sendTelegramMessage(ctx context.Context, chatID, text string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramToken)
	values := url.Values{}
	values.Set("chat_id", chatID)
	values.Set("text", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) CreateLinkToken(ctx context.Context, userID int64, username string) (string, string, error) {
	token, err := generateToken(32)
	if err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	link := ""
	if s.botName != "" {
		link = fmt.Sprintf("https://t.me/%s?start=%s", s.botName, token)
	}
	if err := s.storage.CreateLinkToken(ctx, domain.LinkToken{
		Token:    token,
		UserID:   userID,
		Username: username,
	}); err != nil {
		return "", "", err
	}
	return token, link, nil
}

func (s *Service) BindByToken(ctx context.Context, token, chatID, username string) error {
	link, err := s.storage.ConsumeLinkToken(ctx, token)
	if err != nil {
		return err
	}
	// prefer username from telegram update, fallback to stored
	name := strings.TrimPrefix(username, "@")
	if name == "" {
		name = link.Username
	}
	contact := domain.Contact{
		UserID:   link.UserID,
		Username: name,
		ChatID:   chatID,
	}
	return s.storage.UpsertContact(ctx, contact)
}

func generateToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
