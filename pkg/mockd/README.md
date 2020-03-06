# Generate Mock

## Install gomock

GO111MODULE=on go get github.com/golang/mock/mockgen@latest

## Generate interface mock

```
mockgen -source=../kt/command/types.go -destination=./mock/action_mock.go -package=mock
mockgen -source=../kt/cluster/types.go -destination=./mock/kubernetes_mock.go -package=mock
mockgen -source=../kt/connect/types.go -destination=./mock/connect_mock.go -package=mock
```
