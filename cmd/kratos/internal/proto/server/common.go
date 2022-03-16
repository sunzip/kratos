package server

import (
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
	Imports []string
	// func (d *Data) UserConfGrpc() infopb.UserConfClient
	// key: 	*Data-UserConfGrpc
	// value: 	func (d *Data) UserConfGrpc() infopb.UserConfClient
	Methods map[string]MethodInfo
	// 废弃
	FuncBegin   BlockStatus
	ImportBegin BlockStatus
	// key:name
	Structs map[string]StructInfo
}

type MethodInfo struct {
	MethodDefind string
	StartPos     int
	EndPos       int
}

type StructInfo struct {
	StartPos        int
	LastFieldEndPos int
	EndPos          int
	// name-type
	// oplogCli     infopb.OpLogClient > oplogCli-infopb.OpLogClient
	Fields string
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
