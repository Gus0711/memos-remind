package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"

	"github.com/usememos/memos/internal/profile"
	"github.com/usememos/memos/plugin/notification"
	storepb "github.com/usememos/memos/proto/gen/store"
	apiv1 "github.com/usememos/memos/server/router/api/v1"
	"github.com/usememos/memos/server/router/fileserver"
	"github.com/usememos/memos/server/router/frontend"
	"github.com/usememos/memos/server/router/rss"
	reminderrunner "github.com/usememos/memos/server/runner/reminder"
	"github.com/usememos/memos/server/runner/s3presign"
	"github.com/usememos/memos/store"
)

type Server struct {
	Secret  string
	Profile *profile.Profile
	Store   *store.Store

	echoServer          *echo.Echo
	notificationService *notification.Service
	runnerCancelFuncs   []context.CancelFunc
}

func NewServer(ctx context.Context, profile *profile.Profile, store *store.Store) (*Server, error) {
	fmt.Println("=== NewServer: Starting server initialization ===")

	// Initialize notification service with providers
	notificationService := notification.NewService()
	notificationService.RegisterProvider(notification.NewInboxProvider(store))
	notificationService.RegisterProvider(notification.NewWebPushProvider(store))

	s := &Server{
		Store:               store,
		Profile:             profile,
		notificationService: notificationService,
	}

	echoServer := echo.New()
	echoServer.Debug = true
	echoServer.HideBanner = true
	echoServer.HidePort = true
	echoServer.Use(middleware.Recover())
	s.echoServer = echoServer

	instanceBasicSetting, err := s.getOrUpsertInstanceBasicSetting(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get instance basic setting")
	}

	// Initialize VAPID keys for web push notifications
	fmt.Println("=== NewServer: About to initialize web push setting ===")
	if err := s.initializeWebPushSetting(ctx); err != nil {
		fmt.Printf("=== NewServer: FAILED to initialize web push setting: %v ===\n", err)
	} else {
		fmt.Println("=== NewServer: Web push setting initialization COMPLETED ===")
	}

	secret := "usememos"
	if !profile.Demo {
		secret = instanceBasicSetting.SecretKey
	}
	s.Secret = secret

	// Register healthz endpoint.
	echoServer.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Service ready.")
	})

	// Serve frontend static files.
	frontend.NewFrontendService(profile, store).Serve(ctx, echoServer)

	rootGroup := echoServer.Group("")

	apiV1Service := apiv1.NewAPIV1Service(s.Secret, profile, store)

	// Register HTTP file server routes BEFORE gRPC-Gateway to ensure proper range request handling for Safari.
	// This uses native HTTP serving (http.ServeContent) instead of gRPC for video/audio files.
	fileServerService := fileserver.NewFileServerService(s.Profile, s.Store, s.Secret)
	fileServerService.RegisterRoutes(echoServer)

	// Create and register RSS routes (needs markdown service from apiV1Service).
	rss.NewRSSService(s.Profile, s.Store, apiV1Service.MarkdownService).RegisterRoutes(rootGroup)
	// Register gRPC gateway as api v1.
	if err := apiV1Service.RegisterGateway(ctx, echoServer); err != nil {
		return nil, errors.Wrap(err, "failed to register gRPC gateway")
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	var address, network string
	if len(s.Profile.UNIXSock) == 0 {
		address = fmt.Sprintf("%s:%d", s.Profile.Addr, s.Profile.Port)
		network = "tcp"
	} else {
		address = s.Profile.UNIXSock
		network = "unix"
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}

	// Start Echo server directly (no cmux needed - all traffic is HTTP).
	s.echoServer.Listener = listener
	go func() {
		if err := s.echoServer.Start(address); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start echo server", "error", err)
		}
	}()
	s.StartBackgroundRunners(ctx)

	return nil
}

func (s *Server) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	slog.Info("server shutting down")

	// Cancel all background runners
	for _, cancelFunc := range s.runnerCancelFuncs {
		if cancelFunc != nil {
			cancelFunc()
		}
	}

	// Shutdown echo server.
	if err := s.echoServer.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	// Close database connection.
	if err := s.Store.Close(); err != nil {
		slog.Error("failed to close database", slog.String("error", err.Error()))
	}

	slog.Info("memos stopped properly")
}

func (s *Server) StartBackgroundRunners(ctx context.Context) {
	// Create a separate context for each background runner
	// This allows us to control cancellation for each runner independently
	s3Context, s3Cancel := context.WithCancel(ctx)
	reminderContext, reminderCancel := context.WithCancel(ctx)

	// Store the cancel functions so we can properly shut down runners
	s.runnerCancelFuncs = append(s.runnerCancelFuncs, s3Cancel, reminderCancel)

	// Create and start S3 presign runner
	s3presignRunner := s3presign.NewRunner(s.Store)
	s3presignRunner.RunOnce(ctx)

	// Start continuous S3 presign runner
	go func() {
		s3presignRunner.Run(s3Context)
		slog.Info("s3presign runner stopped")
	}()

	// Create and start reminder runner
	// Checks for due reminders every minute
	reminderRunner := reminderrunner.NewRunner(s.Store, s.notificationService)
	go func() {
		reminderRunner.Run(reminderContext, time.Minute)
		slog.Info("reminder runner stopped")
	}()

	// Log the number of goroutines running
	slog.Info("background runners started", "goroutines", runtime.NumGoroutine())
}

func (s *Server) getOrUpsertInstanceBasicSetting(ctx context.Context) (*storepb.InstanceBasicSetting, error) {
	instanceBasicSetting, err := s.Store.GetInstanceBasicSetting(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get instance basic setting")
	}
	modified := false
	if instanceBasicSetting.SecretKey == "" {
		instanceBasicSetting.SecretKey = uuid.NewString()
		modified = true
	}
	if modified {
		instanceSetting, err := s.Store.UpsertInstanceSetting(ctx, &storepb.InstanceSetting{
			Key:   storepb.InstanceSettingKey_BASIC,
			Value: &storepb.InstanceSetting_BasicSetting{BasicSetting: instanceBasicSetting},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to upsert instance setting")
		}
		instanceBasicSetting = instanceSetting.GetBasicSetting()
	}
	return instanceBasicSetting, nil
}

// initializeWebPushSetting generates VAPID keys if they don't exist.
func (s *Server) initializeWebPushSetting(ctx context.Context) error {
	fmt.Println(">>> initializeWebPushSetting: STARTING")
	webPushSetting, err := s.Store.GetInstanceWebPushSetting(ctx)
	if err != nil {
		fmt.Printf(">>> initializeWebPushSetting: FAILED to get setting: %v\n", err)
		return errors.Wrap(err, "failed to get web push setting")
	}
	fmt.Printf(">>> initializeWebPushSetting: got setting - privateKeyLen=%d, publicKeyLen=%d, enabled=%v\n",
		len(webPushSetting.VapidPrivateKey), len(webPushSetting.VapidPublicKey), webPushSetting.Enabled)

	// If VAPID keys already exist, no need to generate
	if webPushSetting.VapidPrivateKey != "" && webPushSetting.VapidPublicKey != "" {
		fmt.Println(">>> initializeWebPushSetting: VAPID keys already exist, skipping generation")
		return nil
	}

	fmt.Println(">>> initializeWebPushSetting: Generating new VAPID keys...")
	// Generate new VAPID keys
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		fmt.Printf(">>> initializeWebPushSetting: FAILED to generate keys: %v\n", err)
		return errors.Wrap(err, "failed to generate VAPID keys")
	}
	fmt.Printf(">>> initializeWebPushSetting: Generated keys - privateKeyLen=%d, publicKeyLen=%d\n",
		len(privateKey), len(publicKey))

	// Save the new keys (enabled by default)
	webPushSetting.VapidPrivateKey = privateKey
	webPushSetting.VapidPublicKey = publicKey
	webPushSetting.Enabled = true

	fmt.Println(">>> initializeWebPushSetting: Saving to store...")
	_, err = s.Store.UpsertInstanceSetting(ctx, &storepb.InstanceSetting{
		Key:   storepb.InstanceSettingKey_WEB_PUSH,
		Value: &storepb.InstanceSetting_WebPushSetting{WebPushSetting: webPushSetting},
	})
	if err != nil {
		fmt.Printf(">>> initializeWebPushSetting: FAILED to save: %v\n", err)
		return errors.Wrap(err, "failed to save web push setting")
	}

	fmt.Println(">>> initializeWebPushSetting: SUCCESS - VAPID keys are now configured!")
	return nil
}
