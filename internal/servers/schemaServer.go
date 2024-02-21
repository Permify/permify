package servers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/rs/xid"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaServer - Structure for Schema Server
type SchemaServer struct {
	v1.UnimplementedSchemaServer

	sw storage.SchemaWriter
	sr storage.SchemaReader
	su storage.SchemaUpdater
}

// NewSchemaServer - Creates new Schema Server
func NewSchemaServer(sw storage.SchemaWriter, sr storage.SchemaReader, su storage.SchemaUpdater) *SchemaServer {
	return &SchemaServer{
		sw: sw,
		sr: sr,
		su: su,
	}
}

// Write - Configure new Permify Schema to Permify
func (r *SchemaServer) Write(ctx context.Context, request *v1.SchemaWriteRequest) (*v1.SchemaWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.write")
	defer span.End()
	version := xid.New().String()

	cnf, err := parseAndCompileSchema(request.GetSchema(), request.GetTenantId(), version)	
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	err = r.sw.WriteSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaWriteResponse{
		SchemaVersion: version,
	}, nil
}

// Read - Read created Schema
func (r *SchemaServer) Read(ctx context.Context, request *v1.SchemaReadRequest) (*v1.SchemaReadResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.read")
	defer span.End()

	version := request.GetMetadata().GetSchemaVersion()
	if version == "" {
		ver, err := r.sr.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return nil, status.Error(GetStatus(err), err.Error())
		}
		version = ver
	}

	response, err := r.sr.ReadSchema(ctx, request.GetTenantId(), version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.SchemaReadResponse{
		Schema: response,
	}, nil
}

// PartialWrite - Update existing Schema
func (r *SchemaServer) PartialWrite(
	ctx context.Context,
	request *v1.SchemaPartialWriteRequest,
) (*v1.SchemaPartialWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "schemas.partial-write")
	defer span.End()

	version := request.GetMetadata().GetSchemaVersion()
	if version == "" {
		ver, err := r.sr.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return nil, status.Error(GetStatus(err), err.Error())
		}
		version = ver
	}
	// allErrors collects all errors occuring from UpdateSchema
	var allErrors []error // Collect all errors

	parsedEntities := make(map[string]map[string][]string)
	entities := request.GetEntities()
	for entity, schemas := range entities {
		parsedEntities[entity] = make(map[string][]string)
		parseSchema := func(schema string, category string) error {
			sch, err := parser.NewParser(schema).Parse()
			if err != nil {
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				return fmt.Errorf("error parsing schema for entity '%s', category '%s': %w", entity, category, err)
			}
			for _, st := range sch.Statements {
				parsedEntities[entity][category] = append(parsedEntities[entity][category], st.String())
			}
			return nil
		}
		parseSchemas := func(schemas []string, category string) error {
			for _, schema := range schemas {
				if err := parseSchema(schema, category); err != nil {
					return err
				}
			}
			return nil
		}

		// Parse write schemas and append them to parsed entity
		if err := parseSchemas(schemas.GetWrite(), "write"); err != nil {
			allErrors = append(allErrors, status.Error(GetStatus(err), err.Error()))
		}

		// Parse update schemas and append them to parsed entity
		if err := parseSchemas(schemas.GetUpdate(), "update"); err != nil {
			allErrors = append(allErrors, status.Error(GetStatus(err), err.Error()))
		}

		// Parse delete schemas and append them to parsed entity
		if err := parseSchemas(schemas.GetDelete(), "delete"); err != nil {
			allErrors = append(allErrors, status.Error(GetStatus(err), err.Error()))
		}
	}

	// Check for errors and return
	if len(allErrors) > 0 {
		return nil, &MultiError{Errors: allErrors}
	}
	schema, err := r.su.UpdateSchema(ctx, request.GetTenantId(), version, parsedEntities)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}
	
	version = xid.New().String()

	cnf, err := parseAndCompileSchema(strings.Join(schema, "\n"), request.GetTenantId(), version)	
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	err = r.sw.WriteSchema(ctx, cnf)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}	

	return &v1.SchemaPartialWriteResponse{
		SchemaVersion: version,
	}, nil
}

func parseAndCompileSchema(schema, tenantID, version string) ([]storage.SchemaDefinition, error) {
	sch, err := parser.NewParser(schema).Parse()
	if err != nil {
		return nil, err
	}

	_, _, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		return nil, err
	}
	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             tenantID,
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}
	return cnf, nil
}
