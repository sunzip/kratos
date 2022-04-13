package server

import (
	"bytes"
	"html/template"
)

var serviceTemplate = `
{{- /* delete empty line */ -}}
package service

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	"errors"
	{{- if .UseIO }}
	"io"
	{{- end }}

	{{ .GrpcPbName }} "{{ .GrpcPackage }}"
	"git.hiscene.net/hi_uav/uav-command-server/common/tools"
	{{ .HttpPbName }} "{{ .Package }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"{{ .DomainPackage }}"
	"git.hiscene.net/hifoundry/go-kit/util/hiKratos"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/status"
)

type {{ .Service }}Service struct {
	// pb.Unimplemented{{ .Service }}Server
	logger *log.Helper
	repo    domain.I{{ .Service }}Repo
}

func New{{ .Service }}Service(logger log.Logger, repo domain.I{{ .Service }}Repo) domain.I{{ .Service }}Service {
	return &{{ .Service }}Service{
		logger:  log.NewHelper(logger),
		repo:    repo,
	}
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}Service) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*{{ .HttpPbName }}.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*{{ .HttpPbName }}.{{ .Reply }}{{ end }}, error) {
	{{ if eq .Reply $s1 }}return &emptypb.Empty{}{{ else }}
	reqData := &{{ .GrpcPbName }}.{{ .Request }}{}
	err := tools.StructConvert(reqData, req)
	if err != nil {
		return nil, hiKratos.ResponseErr(ctx, {{ .HttpPbName }}.ErrorInvalidParameter)
	}
	repData, err := s.repo.{{ .Name }}(ctx,reqData)
	if err != nil {
		statusErr := status.FromContextError(context.DeadlineExceeded)
		if errors.Is(err, statusErr.Err()) {
			return nil, hiKratos.ResponseErr(ctx, pb.ErrorTimeout)
		} else {
			return nil, hiKratos.ResponseErr(ctx, {{ .HttpPbName }}.ErrorInternalError)
		}
	}
	rep := &{{ .HttpPbName }}.{{ .Reply }}{}
	err = tools.StructConvert(rep, repData)
	if err != nil {
		return nil, hiKratos.ResponseErr(ctx, {{ .HttpPbName }}.ErrorInternalError)
	}
	return rep, err{{ end }}
	// return {{ if eq .Reply $s1 }}&emptypb.Empty{}{{ else }}&{{ .HttpPbName }}.{{ .Reply }}{}{{ end }}, nil
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&{{ .HttpPbName }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 3 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&{{ .HttpPbName }}.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 4 }}
func (s *{{ .Service }}Service) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*{{ .HttpPbName }}.{{ .Request }}{{ end }}, conn {{ .HttpPbName }}.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&{{ .HttpPbName }}.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- end }}
{{- end }}
`

type MethodType uint8

const (
	unaryType          MethodType = 1
	twoWayStreamsType  MethodType = 2
	requestStreamsType MethodType = 3
	returnsStreamsType MethodType = 4
)

type IService interface {
	Execute() ([]byte, error)
	ServiceName() string
}

// Service is a proto service.
type Service struct {
	// import pb 完整package; 本服务的， HTTP proto
	Package string
	// grpc proto package
	GrpcPackage string
	// 包重命名
	HttpPbName string
	// 包重命名
	GrpcPbName string
	// 项目名称, xx/info/api/xx, 这里就是info
	ProjectName string
	// 生成哪一层的代码, http grpc
	Layer string
	// 引用module/app等的包名
	ModulePackage string
	// 引用 domain 包
	DomainPackage string
	// 引用 internal 包, 可以灵活引用internal下面的包
	InternalPackage string
	// 服务名称 OpLog ; proto文件里定义的service
	Service string
	// 服务名称 小写, 用作包名； opLog; 路径里的名称
	ServiceLower string
	Methods      []*Method
	GoogleEmpty  bool

	UseIO      bool
	UseContext bool
	// proto 源文件
	SourceProto string
	// 生成工具名称
	ToolName string
}

// Method is a proto method.
type Method struct {
	// 包重命名
	HttpPbName string
	// 包重命名
	GrpcPbName string
	Service    string
	Name       string
	Request    string
	Reply      string

	// type: unary or stream
	Type MethodType
}

// 根据servicename 获取go文件名称
func (s *Service) ServiceName(layer FileLayer) string {
	if layer == ProviderSetLayer {
		return "provider_set"
	} else {
		return s.Service + layer.GetLayerStr("_", "")
	}
}
func (s *Service) Execute(layer FileLayer) ([]byte, error) {
	switch layer {
	case AppLayer:
		return s.executeApp()
	case ServiceLayer:
		return s.execute()
	case RepoLayer:
		return s.executeRepo()
	case ProviderSetLayer:
		return s.executeProviderSet()
	case DomainLayer:
		return s.executeDomain()
	}
	return s.execute()
}
func (s *Service) execute() ([]byte, error) {
	const empty = "google.protobuf.Empty"
	buf := new(bytes.Buffer)
	for _, method := range s.Methods {
		if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
			(method.Type == returnsStreamsType && method.Request == empty) {
			s.GoogleEmpty = true
		}
		if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
			s.UseIO = true
		}
		if method.Type == unaryType {
			s.UseContext = true
		}
	}
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
