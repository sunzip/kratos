package server

var serviceTemplateGrpc = `
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

	// {{ .GrpcPbName }} "{{ .GrpcPackage }}"
	"git.hiscene.net/hi_uav/uav-command-server/common/tools"
	commonPb "git.hiscene.net/hiar_mozi/server/mozi-common/api/mozi/common/v1"  {{- if eq 2 1 }} 注释,此行不需要跟随项目变动, common里的pb定义,包括err {{ end }}
	{{ .HttpPbName }} "{{ .Package }}" {{- if eq 2 1 }} 注释, pb {{ end }}
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
	"{{ .DomainPackage }}"
	// "git.hiscene.net/hiar_mozi/server/mozi-device-service/internal/pkg/customerErr"//这里可以提到common里
	mozitools "git.hiscene.net/hiar_mozi/server/mozi-common/tools"
	// "git.hiscene.net/hiar_mozi/server/mozi-device-service/internal/pkg/util"//这里可以提到common里
	"git.hiscene.net/hifoundry/go-kit/util/hiKratos"
	kerrors "github.com/sunzip/kratos/v2/errors"
	"github.com/sunzip/kratos/v2/log"
	"google.golang.org/grpc/status"
)

var(
	module="{{ .Service }}.service"
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
	method:="{{ .Name }}"
	{{ if eq .Request $s1 }}{{ if eq .Reply $s1 }}_{{else}}repData{{end}}, err := s.repo.{{ .Name }}(ctx){{ else }}reqData := &{{ .GrpcPbName }}.{{ .Request }}{}
	err := tools.StructConvert(reqData, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("module",module,"method",method,"StructConvert err",err)
		return nil, hiKratos.ResponseErr(ctx, commonPb.ErrorInvalidParameter)
	}
	{{ if eq .Reply $s1 }}{{else}}repData, {{end}}err {{ if eq .Reply $s1 }}{{else}}:{{end}}= s.repo.{{ .Name }}(ctx,reqData){{end}}
	if err != nil {
		statusErr := status.FromContextError(context.DeadlineExceeded)
		if errors.Is(err, statusErr.Err()) || errors.Is(err, context.DeadlineExceeded) {
			s.logger.WithContext(ctx).Errorw("module",module,"method",method,"repo.{{ .Name }} timeout",err)
			return nil, hiKratos.ResponseErr(ctx, commonPb.ErrorTimeout)
		} else if cErr, ok := err.(mozitools.CustomeErr); ok {
			s.logger.WithContext(ctx).Errorw("module", module, "method", method, "repo.{{ .Name }} CustomeErr", err)
			return nil, hiKratos.ResponseErr(ctx, cErr.Error4Kratos)
		} else if _, ok := err.(*kerrors.Error); ok {
			s.logger.WithContext(ctx).Errorw("module", module, "method", method, "repo.{{ .Name }} kratos err", err)
			return nil, err
		} else {
			s.logger.WithContext(ctx).Errorw("module",module,"method",method,"repo.{{ .Name }} InternalError",err)
			return nil, hiKratos.ResponseErr(ctx, commonPb.ErrorInternalError)
		}
	}
	{{ if eq .Reply $s1 }}rep :=&emptypb.Empty{}{{ else }}rep := &{{ .HttpPbName }}.{{ .Reply }}{}
	err = tools.StructConvert(rep, repData)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("module",module,"method",method,"StructConvert err",err)
		return nil, hiKratos.ResponseErr(ctx, commonPb.ErrorInternalError)
	}{{end}}	
	return rep, err
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
