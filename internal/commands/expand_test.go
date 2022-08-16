package commands

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("expand-command", func() {
	// var expandCommand *ExpandCommand
	// var l = logger.New("debug")

	// DRIVE SAMPLE
	//driveConfigs := []entities.EntityConfig{
	//	{
	//		Entity:           "user",
	//		SerializedConfig: []byte("entity user {}"),
	//	},
	//	{
	//		Entity:           "organization",
	//		SerializedConfig: []byte("entity organization {\nrelation admin @user\n}"),
	//	},
	//	{
	//		Entity:           "folder",
	//		SerializedConfig: []byte("entity folder {\n relation\tparent\t@organization\nrelation\tcreator\t@user\nrelation\tcollaborator\t@user\n action read = collaborator\naction update = collaborator\naction delete = creator or parent.admin\n}"),
	//	},
	//	{
	//		Entity:           "doc",
	//		SerializedConfig: []byte("entity doc {\nelation\tparent\t@organization\nrelation\towner\t@user\n  action read = (owner or parent.collaborator) or parent.admin\naction update = owner and parent.admin\n action delete = owner or parent.admin\n}"),
	//	},
	//}

	Context("Drive Sample: Expand", func() {
	})
})
