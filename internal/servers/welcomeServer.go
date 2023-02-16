package servers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Permify/permify/internal"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// WelcomeServer - Structure for Welcome Server
type WelcomeServer struct {
	v1.UnimplementedWelcomeServer
}

// NewWelcomeServer - Creates new Welcome Server
func NewWelcomeServer() *WelcomeServer {
	return &WelcomeServer{}
}

func (r *WelcomeServer) Hello(context.Context, *emptypb.Empty) (*v1.WelcomeResponse, error) {
	return &v1.WelcomeResponse{
		Permify: internal.OneLiner,
		Sources: &v1.WelcomeResponse_Sources{
			Docs:   internal.Docs,
			GitHub: internal.GitHub,
			Blog:   internal.Blog,
		},
		Socials: &v1.WelcomeResponse_Socials{
			Discord:  internal.Discord,
			Twitter:  internal.Twitter,
			Linkedin: internal.Linkedin,
		},
	}, nil
}
