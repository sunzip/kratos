package client

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

// 为错误码枚举的翻译做解析
// 在当前路径下 dir/i18n/example.{zh}.toml
func do4ErrorToml(protoFile string) {
	if !strings.HasSuffix(protoFile, "errors.proto") {
		return
	}
	dir := filepath.Dir(protoFile)
	reader, err := os.Open(protoFile)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var ()
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {

		}),
		proto.WithEnum(func(o *proto.Enum) {
			buf := new(bytes.Buffer)
			lan := "zh"
			for _, item := range o.Elements {
				if enumField, ok := item.(*proto.EnumField); ok {
					var langName string
					fmt.Println(enumField)
					if enumField.Comment != nil {
						l := len(enumField.Comment.Lines)
						if l > 0 {
							// 支持多语言, 可以根据comment的行, 也可以解析最后一行
							langName = enumField.Comment.Lines[l-1]
						}
					}
					if len(langName) == 0 {
						log.Fatalf("错误码%s没有翻译", enumField.Name)
					}
					buf.WriteString(fmt.Sprintf("%-40s = \"%s\"\n", enumField.Name, langName))
				}

			}

			to := filepath.Join(dir, "i18n", fmt.Sprintf("example.%s.toml", lan))
			if err := os.WriteFile(to, buf.Bytes(), 0o644); err != nil {
				log.Fatal(err)
			}
		}),
		proto.WithService(func(s *proto.Service) {

			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if ok {
					fmt.Println(r)
				}

			}
		}),
	)
}
