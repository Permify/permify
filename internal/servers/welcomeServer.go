package servers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

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
		Permify: "Open-source authorization service inspired by Google Zanzibar",
		Sources: &v1.WelcomeResponse_Sources{
			Docs:   "https://github.com/Permify/permify",
			GitHub: "https://github.com/Permify/permify",
			Blog:   "https://github.com/Permify/permify",
		},
		Socials: &v1.WelcomeResponse_Socials{
			Discord:  "https://github.com/Permify/permify",
			Twitter:  "https://github.com/Permify/permify",
			Linkedin: "https://github.com/Permify/permify",
		},
	}, nil
}
