module github.com/sunzip/kratos/contrib/registry/nacos/v2

go 1.16

require (
	github.com/sunzip/kratos/v2 v2.2.0
	github.com/nacos-group/nacos-sdk-go v1.0.9
)

replace github.com/sunzip/kratos/v2 => ../../../

replace github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
