module github.com/sunzip/kratos/contrib/config/nacos/v2

go 1.16

require (
	github.com/sunzip/kratos/v2 v2.2.0
	github.com/nacos-group/nacos-sdk-go v1.0.9
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/sunzip/kratos/v2 => ../../../

replace github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
