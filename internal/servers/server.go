package servers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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

	"github.com/Permify/permify/internal/authn"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/servers/middleware"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	grpcV1 "github.com/Permify/permify/pkg/pb/base/v1"
)

var tracer = otel.Tracer("servers")

// ServiceContainer -
type ServiceContainer struct {
	RelationshipService services.IRelationshipService
	PermissionService   services.IPermissionService
	SchemaService       services.ISchemaService
}

// Run -
func (s *ServiceContainer) Run(ctx context.Context, cfg *config.Server, authentication *config.Authn, l *logger.Logger) error {
	var err error

	interceptors := []grpc.UnaryServerInterceptor{
		grpcValidator.UnaryServerInterceptor(),
	}

	var authenticator authn.KeyAuthenticator
	if authentication != nil && authentication.Enabled {
		authenticator, err = authn.NewKeyAuthn(authentication.Keys...)
		interceptors = append(interceptors, grpcAuth.UnaryServerInterceptor(middleware.AuthFunc(authenticator)))
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
	}

	if cfg.GRPC.TLSConfig.Enabled {
		var c credentials.TransportCredentials
		c, err = credentials.NewServerTLSFromFile(cfg.GRPC.TLSConfig.CertPath, cfg.GRPC.TLSConfig.KeyPath)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(c))
	}

	grpcServer := grpc.NewServer(opts...)
	grpcV1.RegisterPermissionServer(grpcServer, NewPermissionServer(s.PermissionService, l))
	grpcV1.RegisterSchemaServer(grpcServer, NewSchemaServer(s.SchemaService, l))
	grpcV1.RegisterRelationshipServer(grpcServer, NewRelationshipServer(s.RelationshipService, l))
	health.RegisterHealthServer(grpcServer, NewHealthServer())
	grpcV1.RegisterWelcomeServer(grpcServer, NewWelcomeServer())
	reflection.Register(grpcServer)

	var lis net.Listener
	lis, err = net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		return err
	}

	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			l.Error("failed to start grpc server", err)
		}
	}()

	l.Info(fmt.Sprintf("🚀 grpc server successfully started: %s", cfg.GRPC.Port))

	var httpServer *http.Server
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
		defer conn.Close()

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
		if err = grpcV1.RegisterWelcomeHandler(ctx, mux, conn); err != nil {
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

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(ctx); err != nil {
			l.Error(err)
			return err
		}
	}

	grpcServer.GracefulStop()

	l.Info("gracefully shutting down")

	return nil
}
