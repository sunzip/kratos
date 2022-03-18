package server

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
)

var (
	// pb的别名
	HttpPbName = "pb"
	// pb的别名
	GrpcPbName = "infoPb"
)

// CmdServer the service command.
var CmdServer = &cobra.Command{
	Use:   "server",
	Short: "Generate the proto Server implementations",
	Long:  "Generate the proto Server implementations. Example: kratos proto server api/xxx.proto -target-dir=internal/service",
	Run:   run,
}
var targetDir string

func init() {
	CmdServer.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/service", "generate target directory")
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto server api/xxx.proto")
		return
	}
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var (
		pkg string
		res []*Service
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				pkg = strings.Split(o.Constant.Source, ";")[0]
			}
		}),
		proto.WithService(func(s *proto.Service) {
			cs := &Service{
				Package: pkg,
				Service: s.Name,
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if ok {
					cs.Methods = append(cs.Methods, &Method{
						GrpcPbName: GrpcPbName,
						HttpPbName: HttpPbName,
						Service:    s.Name, Name: r.Name, Request: r.RequestType,
						Reply: r.ReturnsType, Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
					})
				}
			}
			cs.SourceProto = args[0]

			pathParts := strings.Split(targetDir, "/")
			cs.ServiceLower = pathParts[len(pathParts)-1] //strings.ToLower(cs.Service)
			cs.ToolName = "createService"
			cs.HttpPbName = HttpPbName
			cs.GrpcPbName = GrpcPbName
			cs.GrpcPackage = changeDir(cs.Package, "/drone-appservice/", "/common/api/drone/info/v1")
			res = append(res, cs)
		}),
	)
	checkServiceDir(targetDir)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exsits\n", targetDir)
		return
	}
	for _, s := range res {
		{
			// data.go文件增加Data类字段,Data类方法,NewData函数增加代码

			// ok
			dataPath := changeDir(targetDir, "/module/", "/data")
			sf := parseFile(dataPath, "data.go", s)
			incrementMethodData(dataPath, "data.go", sf, s)

			// ok
			protocolPath := changeDir(targetDir, "/module/", "/protocol")
			sfReg := parseFile(protocolPath, "register.go", s)
			incrementRegister(protocolPath, "register.go", sfReg, s)

			cmdServerPath := changeDir(targetDir, "/internal/", "/cmd/server")
			sfCmdServer := parseFile(cmdServerPath, "wire.go", s)
			incrementCmdServer(cmdServerPath, "wire.go", sfCmdServer, s)
			// return
		}

		s.ModulePackage = changeDir(s.Package, "/api/", "/internal/module")
		s.DomainPackage = changeDir(s.Package, "/api/", "/internal/domain")
		s.InternalPackage = changeDir(s.Package, "/api/", "/internal")

		GenerateFile(targetDir, s, AppLayer)
		GenerateFile(targetDir, s, ServiceLayer)
		GenerateFile(targetDir, s, RepoLayer)
		GenerateFile(targetDir, s, ProviderSetLayer)

		domainDir := changeDir(targetDir, "/module/", "")
		GenerateFile(domainDir, s, DomainLayer)
	}
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}
func checkServiceDir(targetDir string) {
	// targetDir 为服务的根目录,
	checkDir(targetDir)
	checkDir(targetDir + "/service")
	checkDir(targetDir + "/app")
	checkDir(targetDir + "/repo")
}

// 不存在就创建
func checkDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//
		os.Mkdir(dir, os.ModePerm)
	}
}

type FileLayer uint8

const (
	AppLayer FileLayer = iota
	ServiceLayer
	RepoLayer
	ProviderSetLayer
	DomainLayer
)

// 比如app, prefix=_ ,则返回_app
//  返回后缀 xxx_app.go
//  返回层名称, 和前后缀
func (t FileLayer) GetLayerStr(prefix, suffix string) (ret string) {
	switch t {
	case AppLayer:
		ret = prefix + "app" + suffix
	case ServiceLayer:
		ret = prefix + "service" + suffix
	case RepoLayer:
		ret = prefix + "repo" + suffix
	case ProviderSetLayer:
		ret = ""
	case DomainLayer:
		ret = prefix + "domain" + suffix
	}
	return
}

func GenerateFile(targetDir string, s *Service, layer FileLayer) {

	to := path.Join(targetDir, layer.GetLayerStr("", ""), strings.ToLower(s.ServiceName(layer))+".go")

	if _, err := os.Stat(to); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s already exists: %s\n", s.ServiceName(layer), to)
		return
	}
	b, err := s.Execute(layer)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(to)
}
