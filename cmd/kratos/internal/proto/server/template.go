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
	{{- if .UseIO }}
	"io"
	{{- end }}

	pb "{{ .Package }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"{{ .DomainPackage }}"
	"github.com/go-kratos/kratos/v2/log"
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
func (s *{{ .Service }}Service) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return {{ if eq .Reply $s1 }}&emptypb.Empty{}{{ else }}s.repo.{{ .Name }}(ctx,req){{ end }}
	// return {{ if eq .Reply $s1 }}&emptypb.Empty{}{{ else }}&pb.{{ .Reply }}{}{{ end }}, nil
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 3 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&pb.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 4 }}
func (s *{{ .Service }}Service) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}, conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&pb.{{ .Reply }}{})
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
	// import pb 完整package
	Package string
	// 引用module/app等的包名
	ModulePackage string
	// 引用 domain 包
	DomainPackage string
	// 引用 internal 包, 可以灵活引用internal下面的包
	InternalPackage string
	// 服务名称 OpLog
	Service string
	// 服务名称 小写, 用作包名
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
	Service string
	Name    string
	Request string
	Reply   string

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
