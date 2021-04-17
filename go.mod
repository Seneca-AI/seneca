module seneca

go 1.16

require (
	cloud.google.com/go/datastore v1.1.0
	cloud.google.com/go/logging v1.3.0
	cloud.google.com/go/storage v1.14.0
	github.com/barasher/go-exiftool v1.3.2
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/go-cmp v0.5.5
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	golang.org/x/oauth2 v0.0.0-20210218202405-ba52d332ba99
	google.golang.org/api v0.40.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/barasher/go-exiftool => github.com/Seneca-AI/go-exiftool v1.3.3-0.20210320190943-82dd1fcee7e3
