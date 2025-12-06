package server

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	queuev1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/queue"
	authclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/auth"
	notifyclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/notification"
	queueclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/queue"
	"github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/middleware"
	"google.golang.org/grpc/status"
)

type Server struct {
	app       *fiber.App
	auth      *authclient.Client
	queue     *queueclient.Client
	notif     *notifyclient.Client
	log       *slog.Logger
	validator *validator.Validate
	appID     int
}

func New(log *slog.Logger, auth *authclient.Client, queue *queueclient.Client, notif *notifyclient.Client, appSecret string, appID int) *Server {
	s := &Server{
		app: fiber.New(fiber.Config{
			AppName: "qflow-api-gateway",
		}),
		auth:      auth,
		queue:     queue,
		notif:     notif,
		log:       log,
		validator: validator.New(validator.WithRequiredStructEnabled()),
		appID:     appID,
	}

	s.app.Use(recover.New())
	s.app.Use(logger.New())
	s.app.Use(cors.New())

	s.routes(appSecret)

	return s
}

func (s *Server) routes(appSecret string) {
	s.app.Post("/auth/register", s.handleRegister)
	s.app.Post("/auth/login", s.handleLogin)
	s.app.Post("/telegram/webhook", s.handleTelegramWebhook)

	// Protected routes
	authMW := middleware.Auth(appSecret)

	s.app.Post("/profile/contact", authMW, s.handleSetContact)
	s.app.Post("/profile/contact/link", authMW, s.handleCreateLinkToken)

	s.app.Get("/queues", authMW, s.handleListQueues)
	s.app.Post("/queues", authMW, s.handleCreateQueue)
	s.app.Get("/queues/:id", authMW, s.handleGetQueue)
	s.app.Post("/queues/:id/join", authMW, s.handleJoinQueue)
	s.app.Post("/queues/:id/leave", authMW, s.handleLeaveQueue)
	s.app.Post("/queues/:id/advance", authMW, s.handleAdvanceQueue)
	s.app.Post("/queues/:id/remove", authMW, s.handleRemoveParticipant)
	s.app.Post("/queues/:id/archive", authMW, s.handleArchiveQueue)
	s.app.Delete("/queues/:id", authMW, s.handleDeleteQueue)
}

func (s *Server) Run(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

type (
	registerReq struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	loginReq struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	contactReq struct {
		TelegramUsername string `json:"telegram_username"`
		ChatID           string `json:"chat_id"`
	}

	createQueueReq struct {
		Title       string `json:"title" validate:"required"`
		Description string `json:"description"`
		Mode        string `json:"mode" validate:"required"`
		GroupCode   string `json:"group_code" validate:"required"`
	}

	groupReq struct {
		GroupCode string `json:"group_code" validate:"required"`
	}

	joinReq struct {
		GroupCode string `json:"group_code" validate:"required"`
		SlotTime  string `json:"slot_time"`
	}

	removeReq struct {
		GroupCode string `json:"group_code" validate:"required"`
		UserID    int64  `json:"user_id" validate:"required,gt=0"`
	}

	linkReq struct {
		TelegramUsername string `json:"telegram_username"`
	}

	telegramUpdate struct {
		Message *struct {
			Text string `json:"text"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			From *struct {
				Username string `json:"username"`
			} `json:"from"`
		} `json:"message"`
	}
)

func (s *Server) handleRegister(c *fiber.Ctx) error {
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id, err := s.auth.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"user_id": id}})
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := s.auth.Login(c.Context(), req.Email, req.Password, int32(s.appID))
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"token": token}})
}

func (s *Server) handleSetContact(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req contactReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.notif.SetContact(c.Context(), user.ID, req.TelegramUsername, req.ChatID); err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": "ok"})
}

func (s *Server) handleCreateLinkToken(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req linkReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	token, link, err := s.notif.CreateLinkToken(c.Context(), user.ID, req.TelegramUsername)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"token": token, "link": link}})
}

func (s *Server) handleTelegramWebhook(c *fiber.Ctx) error {
	var upd telegramUpdate
	if err := c.BodyParser(&upd); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid update")
	}
	if upd.Message == nil {
		return c.SendStatus(fiber.StatusOK)
	}
	token := extractStartToken(upd.Message.Text)
	if token == "" {
		return c.SendStatus(fiber.StatusOK)
	}
	chatID := fmt.Sprintf("%d", upd.Message.Chat.ID)
	username := ""
	if upd.Message.From != nil {
		username = upd.Message.From.Username
	}
	if err := s.notif.BindByToken(c.Context(), token, chatID, username); err != nil {
		s.log.Warn("failed to bind telegram token", slog.Any("err", err))
	}
	return c.SendStatus(fiber.StatusOK)
}

func (s *Server) handleListQueues(c *fiber.Ctx) error {
	group := c.Query("group")
	if group == "" {
		return fiber.NewError(fiber.StatusBadRequest, "group is required")
	}
	queues, err := s.queue.List(c.Context(), group)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": queues})
}

func (s *Server) handleCreateQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var req createQueueReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	mode := parseMode(req.Mode)
	dto, err := s.queue.Create(c.Context(), &queuev1.CreateQueueRequest{
		Title:       req.Title,
		Description: req.Description,
		Mode:        mode,
		GroupCode:   req.GroupCode,
		OwnerId:     user.ID,
	})
	if err != nil {
		return s.mapError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": dto})
}

func (s *Server) handleGetQueue(c *fiber.Ctx) error {
	group := c.Query("group")
	if group == "" {
		return fiber.NewError(fiber.StatusBadRequest, "group is required")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	resp, err := s.queue.Get(c.Context(), id, group)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (s *Server) handleJoinQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req joinReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	pos, err := s.queue.Join(c.Context(), id, user.ID, req.GroupCode, req.SlotTime)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"position": pos}})
}

func (s *Server) handleLeaveQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req groupReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := s.queue.Leave(c.Context(), id, user.ID, req.GroupCode); err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": "ok"})
}

func (s *Server) handleAdvanceQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req groupReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	removed, err := s.queue.Advance(c.Context(), id, user.ID, req.GroupCode)
	if err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": removed})
}

func (s *Server) handleRemoveParticipant(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req removeReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := s.queue.Remove(c.Context(), id, req.UserID, user.ID, req.GroupCode); err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": "ok"})
}

func (s *Server) handleArchiveQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req groupReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := s.queue.Archive(c.Context(), id, user.ID, req.GroupCode); err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": "ok"})
}

func (s *Server) handleDeleteQueue(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req groupReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := s.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := s.queue.Delete(c.Context(), id, user.ID, req.GroupCode); err != nil {
		return s.mapError(err)
	}
	return c.JSON(fiber.Map{"data": "ok"})
}

func (s *Server) mapError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case 3: // InvalidArgument
			return fiber.NewError(fiber.StatusBadRequest, st.Message())
		case 7: // PermissionDenied
			return fiber.NewError(fiber.StatusForbidden, st.Message())
		case 16: // Unauthenticated
			return fiber.NewError(fiber.StatusUnauthorized, st.Message())
		case 6: // AlreadyExists
			return fiber.NewError(fiber.StatusConflict, st.Message())
		case 5: // NotFound
			return fiber.NewError(fiber.StatusNotFound, st.Message())
		case 9: // FailedPrecondition
			return fiber.NewError(fiber.StatusPreconditionFailed, st.Message())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, st.Message())
		}
	}
	return fiber.NewError(fiber.StatusInternalServerError, "internal error")
}

func parseMode(mode string) queuev1.QueueMode {
	switch mode {
	case "live":
		return queuev1.QueueMode_QUEUE_MODE_LIVE
	case "managed":
		return queuev1.QueueMode_QUEUE_MODE_MANAGED
	case "random":
		return queuev1.QueueMode_QUEUE_MODE_RANDOM
	case "slots":
		return queuev1.QueueMode_QUEUE_MODE_SLOTS
	default:
		return queuev1.QueueMode_QUEUE_MODE_LIVE
	}
}

func extractStartToken(text string) string {
	if text == "" {
		return ""
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	if parts[0] != "/start" {
		return ""
	}
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
