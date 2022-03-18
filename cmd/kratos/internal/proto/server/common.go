package server

import (
	"go/ast"
	"runtime"
	"strings"
)

// changeDir(s.Package,"/api/","/internal/module")
//  hi_uav/uav-command-server/common/api/drone/common/v1 -> hi_uav/uav-command-server/common/internal/module
//  将api及以后的路径替换为新路径
//  package /
//  dir \\
func changeDir(dir, replaceDir, newDir string) string {
	pos := strings.Index(dir, replaceDir)

	return dir[:pos] + newDir
}

func parseSourceFile() {

}

type SourceFile struct {
	// wire 使用
	Imports                        LabelIfno
	AutoWireAppPanicBuildLastParam LabelIfno
	// func (d *Data) UserConfGrpc() infopb.UserConfClient
	// key: 	*Data-UserConfGrpc
	// value: 	func (d *Data) UserConfGrpc() infopb.UserConfClient
	Methods map[string]MethodInfo
	// 废弃
	FuncBegin   BlockStatus
	ImportBegin BlockStatus
	// key:name
	Structs map[string]StructInfo
	// key: 备注的第一个词，空格分隔
	//  没用， 注释需要放到语句上面，或者语句后面
	Comments map[string]CommentIfno
	Labels   map[string]LabelIfno
	// key这些语句是否存在，默认不存在
	LineExist map[string]bool
}
type LabelIfno struct {
	Text     string
	StartPos int
	EndPos   int
}
type CommentIfno struct {
	Comment  string
	StartPos int
	EndPos   int
}
type MethodInfo struct {
	MethodDefind string
	StartPos     int
	EndPos       int
}

type StructInfo struct {
	StartPos int
	// 最后一个字段结束的位置
	LastFieldEndPos int
	EndPos          int
	// key:name-type
	// oplogCli     infopb.OpLogClient > oplogCli-infopb.OpLogClient
	Fields map[string]bool
}

type SourceType uint8

const (
	DefaultSourceType SourceType = iota
	ImportSourceType
	FuncSourceType
)

type BlockStatus uint8

const (
	BlockInit BlockStatus = iota
	BlockStart
	BlockEnd
)

func getEndOfLine() string {
	switch runtime.GOOS {
	case "darwin", "linux":
		return "\n"
	case "windows":
		return "\r\n"
	default:
		return "\r\n"
	}
}

// 获取语句表达式的参数列表
func GetExprPara(expr ast.Expr) []ast.Expr {
	call := expr.(*ast.CallExpr)

	return call.Args
}
