package servers

import (
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

type WatchServer struct {
	v1.UnimplementedWatchServer

	dr storage.DataReader
	w  storage.Watcher
}

func NewWatchServer(
	w storage.Watcher,
	dr storage.DataReader,
) *WatchServer {
	return &WatchServer{
		w:  w,
		dr: dr,
	}
}

// Watch function sets up a stream for the client to receive changes.
func (r *WatchServer) Watch(request *v1.WatchRequest, server v1.Watch_WatchServer) error {
	// Start a new context and span for tracing.
	ctx, span := internal.Tracer.Start(server.Context(), "watch.watch")
	defer span.End() // Ensure the span ends when the function returns.

	// Validate the incoming request.
	v := request.Validate()
	if v != nil {
		return v // Return validation error, if any.
	}

	// Extract the snapshot token from the request.
	snap := request.GetSnapToken()
	if snap == "" {
		// If the snapshot token is not provided, get the head snapshot from the database.
		st, err := r.dr.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return err // If there's an error retrieving the snapshot, return it.
		}
		// Encode the snapshot to a string.
		snap = st.Encode().String()
	}

	// Call the Watch function on the watcher, which returns two channels.
	changes, errs := r.w.Watch(ctx, request.GetTenantId(), snap)

	// Create a separate goroutine to handle sending changes to the server.
	go func() {
		for change := range changes {
			// For each change, send it to the client.
			if err := server.Send(&v1.WatchResponse{Changes: change}); err != nil {
				// If an error occurs while sending, exit the goroutine.
				// The error is not handled here because the context will be cancelled if an error is detected.
				return
			}
		}
	}()

	// Main loop for handling errors.
	for err := range errs {
		// If an error occurs, convert it to a status error and return it.
		// This ends the Watch function, which in turn closes the changes channel and ends the above goroutine.
		return status.Error(GetStatus(err), err.Error())
	}

	// At this point, the errs channel has been closed, indicating that no more errors will be coming in.
	// Therefore, it's safe to return nil indicating that the operation was successful.
	return nil
}
