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
			Docs:   "https://docs.permify.co/",
			GitHub: "https://github.com/Permify/permify",
			Blog:   "https://www.permify.co/blog",
		},
		Socials: &v1.WelcomeResponse_Socials{
			Discord:  "https://discord.gg/JJnMeCh6qP",
			Twitter:  "https://twitter.com/GetPermify",
			Linkedin: "https://twitter.com/GetPermify",
		},
	}, nil
}
