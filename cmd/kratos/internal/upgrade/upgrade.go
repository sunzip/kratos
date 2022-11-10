package upgrade

import (
	"fmt"

	"github.com/sunzip/kratos/cmd/kratos/v2/internal/base"

	"github.com/spf13/cobra"
)

// CmdUpgrade represents the upgrade command.
var CmdUpgrade = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the kratos tools",
	Long:  "Upgrade the kratos tools. Example: kratos upgrade",
	Run:   Run,
}

// Run upgrade the kratos tools.
func Run(cmd *cobra.Command, args []string) {
	err := base.GoInstall(
		"github.com/sunzip/kratos/cmd/kratos/v2",
		"github.com/sunzip/kratos/cmd/protoc-gen-go-http/v2",
		"github.com/sunzip/kratos/cmd/protoc-gen-go-errors/v2",
		"google.golang.org/protobuf/cmd/protoc-gen-go",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc",
		"github.com/envoyproxy/protoc-gen-validate",
		"github.com/google/gnostic",
		"github.com/google/gnostic/cmd/protoc-gen-openapi@v0.6.2",
	)
	if err != nil {
		fmt.Println(err)
	}
}
