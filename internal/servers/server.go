package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	health "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/Permify/permify/internal/authn/oidc"
	"github.com/Permify/permify/internal/authn/preshared"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/middleware"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/logger"
	grpcV1 "github.com/Permify/permify/pkg/pb/base/v1"
)

var tracer = otel.Tracer("servers")

// Container is a struct that holds the invoker and various storage storage
// for permission-related operations. It serves as a central point of access
// for interacting with the underlying data and services.
type Container struct {
	// Invoker for performing permission-related operations
	Invoker invoke.Invoker
	// RelationshipReader for reading relationships from storage
	RR storage.RelationshipReader
	// RelationshipWriter for writing relationships to storage
	RW storage.RelationshipWriter
	// SchemaReader for reading schemas from storage
	SR storage.SchemaReader
	// SchemaWriter for writing schemas to storage
	SW storage.SchemaWriter
	// TenantReader for reading tenant information from storage
	TR storage.TenantReader
	// TenantWriter for writing tenant information to storage
	TW storage.TenantWriter

	W storage.Watcher
}

// NewContainer is a constructor for the Container struct.
// It takes an Invoker, RelationshipReader, RelationshipWriter, SchemaReader, SchemaWriter,
// TenantReader, and TenantWriter as arguments, and returns a pointer to a Container instance.
func NewContainer(
	invoker invoke.Invoker,
	rr storage.RelationshipReader,
	rw storage.RelationshipWriter,
	sr storage.SchemaReader,
	sw storage.SchemaWriter,
	tr storage.TenantReader,
	tw storage.TenantWriter,
	w storage.Watcher,
) *Container {
	return &Container{
		Invoker: invoker,
		RR:      rr,
		RW:      rw,
		SR:      sr,
		SW:      sw,
		TR:      tr,
		TW:      tw,
		W:       w,
	}
}

// Run is a method that starts the Container and its services, including the gRPC server,
// an optional HTTP server, and an optional profiler server. It also sets up authentication,
// TLS configurations, and interceptors as needed.
func (s *Container) Run(
	ctx context.Context,
	cfg *config.Server,
	authentication *config.Authn,
	profiler *config.Profiler,
	l *logger.Logger,
) error {
	var err error

	limiter := middleware.NewRateLimiter(cfg.RateLimit) // for example 1000 req/sec

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpcValidator.UnaryServerInterceptor(),
		grpcRecovery.UnaryServerInterceptor(),
		ratelimit.UnaryServerInterceptor(limiter),
	}

	streamingInterceptors := []grpc.StreamServerInterceptor{
		grpcValidator.StreamServerInterceptor(),
		grpcRecovery.StreamServerInterceptor(),
		ratelimit.StreamServerInterceptor(limiter),
	}

	// Configure authentication based on the provided method ("preshared" or "oidc").
	// Add the appropriate interceptors to the unary and streaming interceptors.
	if authentication != nil && authentication.Enabled {
		switch authentication.Method {
		case "preshared":
			var authenticator *preshared.KeyAuthn
			authenticator, err = preshared.NewKeyAuthn(ctx, authentication.Preshared)
			if err != nil {
				return err
			}
			unaryInterceptors = append(unaryInterceptors, grpcAuth.UnaryServerInterceptor(middleware.KeyAuthFunc(authenticator)))
			streamingInterceptors = append(streamingInterceptors, grpcAuth.StreamServerInterceptor(middleware.KeyAuthFunc(authenticator)))
		case "oidc":
			var authenticator *oidc.Authn
			authenticator, err = oidc.NewOidcAuthn(ctx, authentication.Oidc)
			if err != nil {
				return err
			}
			unaryInterceptors = append(unaryInterceptors, oidc.UnaryServerInterceptor(authenticator))
			streamingInterceptors = append(streamingInterceptors, oidc.StreamServerInterceptor(authenticator))
		default:
			return fmt.Errorf("unkown authentication method: '%s'", authentication.Method)
		}
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamingInterceptors...),
	}

	if cfg.GRPC.TLSConfig.Enabled {
		var c credentials.TransportCredentials
		c, err = credentials.NewServerTLSFromFile(cfg.GRPC.TLSConfig.CertPath, cfg.GRPC.TLSConfig.KeyPath)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(c))
	}

	// Create a new gRPC server with the configured interceptors and optional TLS credentials.
	// Register the various service implementations.
	grpcServer := grpc.NewServer(opts...)
	grpcV1.RegisterPermissionServer(grpcServer, NewPermissionServer(s.Invoker, l))
	grpcV1.RegisterSchemaServer(grpcServer, NewSchemaServer(s.SW, s.SR, l))
	grpcV1.RegisterRelationshipServer(grpcServer, NewRelationshipServer(s.RR, s.RW, s.SR, l))
	grpcV1.RegisterTenancyServer(grpcServer, NewTenancyServer(s.TR, s.TW, l))
	grpcV1.RegisterWatchServer(grpcServer, NewWatchServer(s.W, s.RR, l))
	health.RegisterHealthServer(grpcServer, NewHealthServer())
	reflection.Register(grpcServer)

	// Start the profiler server if enabled.
	if profiler.Enabled {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		go func() {
			l.Info(fmt.Sprintf("🚀 profiler server successfully started: %s", profiler.Port))

			pprofserver := &http.Server{
				Addr:         ":" + profiler.Port,
				Handler:      mux,
				ReadTimeout:  20 * time.Second,
				WriteTimeout: 20 * time.Second,
				IdleTimeout:  15 * time.Second,
			}

			if err = pprofserver.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					l.Fatal("failed to start profiler", err)
				}
			}
		}()
	}

	var lis net.Listener
	lis, err = net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		return err
	}

	// Start the gRPC server.
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			l.Error("failed to start grpc server", err)
		}
	}()

	l.Info(fmt.Sprintf("🚀 grpc server successfully started: %s", cfg.GRPC.Port))

	var httpServer *http.Server

	// Start the optional HTTP server with CORS and optional TLS configurations.
	// Connect to the gRPC server and register the HTTP handlers for each service.
	if cfg.HTTP.Enabled {
		options := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		}
		if cfg.GRPC.TLSConfig.Enabled {
			c, err := credentials.NewClientTLSFromFile(cfg.GRPC.TLSConfig.CertPath, "")
			if err != nil {
				return err
			}
			options = append(options, grpc.WithTransportCredentials(c))
		} else {
			options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(timeoutCtx, ":"+cfg.GRPC.Port, options...)
		if err != nil {
			return err
		}
		defer func() {
			if err = conn.Close(); err != nil {
				l.Fatal("Failed to close gRPC connection: %v", err)
			}
		}()

		healthClient := health.NewHealthClient(conn)
		muxOpts := []runtime.ServeMuxOption{
			runtime.WithHealthzEndpoint(healthClient),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
				Marshaler: &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: true,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				},
			}),
		}

		mux := runtime.NewServeMux(muxOpts...)

		if err = grpcV1.RegisterPermissionHandler(ctx, mux, conn); err != nil {
			return err
		}
		if err = grpcV1.RegisterSchemaHandler(ctx, mux, conn); err != nil {
			return err
		}
		if err = grpcV1.RegisterRelationshipHandler(ctx, mux, conn); err != nil {
			return err
		}
		if err = grpcV1.RegisterTenancyHandler(ctx, mux, conn); err != nil {
			return err
		}

		httpServer = &http.Server{
			Addr: ":" + cfg.HTTP.Port,
			Handler: cors.New(cors.Options{
				AllowCredentials: true,
				AllowedOrigins:   cfg.HTTP.CORSAllowedOrigins,
				AllowedHeaders:   cfg.HTTP.CORSAllowedHeaders,
				AllowedMethods: []string{
					http.MethodGet, http.MethodPost,
					http.MethodHead, http.MethodPatch, http.MethodDelete, http.MethodPut,
				},
			}).Handler(mux),
			ReadHeaderTimeout: 5 * time.Second,
		}

		// Start the HTTP server with TLS if enabled, otherwise without TLS.
		go func() {
			var err error
			if cfg.HTTP.TLSConfig.Enabled {
				err = httpServer.ListenAndServeTLS(cfg.HTTP.TLSConfig.CertPath, cfg.HTTP.TLSConfig.KeyPath)
			} else {
				err = httpServer.ListenAndServe()
			}
			if err != http.ErrServerClosed {
				l.Error(err)
			}
		}()

		l.Info(fmt.Sprintf("🚀 http server successfully started: %s", cfg.HTTP.Port))
	}

	// Wait for the context to be canceled (e.g., due to a signal).
	<-ctx.Done()

	// Shutdown the servers gracefully.
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(ctxShutdown); err != nil {
			l.Error(err)
			return err
		}
	}

	// Gracefully stop the gRPC server.
	grpcServer.GracefulStop()

	l.Info("gracefully shutting down")

	return nil
}
