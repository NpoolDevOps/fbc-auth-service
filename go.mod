module github.com/NpoolDevOps/fbc-auth-service

go 1.15

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210210082337-af230fd03ce7
	github.com/NpoolDevOps/fbc-devops-service v0.0.0-20210318090754-3e9f35202388
	github.com/NpoolDevOps/fbc-license-service v0.0.0-20210318060048-fc9b7662d874
	github.com/NpoolRD/http-daemon v0.0.0-20210210091512-241ac31803ef
	github.com/google/uuid v1.2.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)
