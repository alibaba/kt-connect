# Generate Mock

## Install gomock

GO111MODULE=on go get github.com/golang/mock/mockgen@latest

## Generate interface mock

```
mockgen -source=../kt/command/types.go -destination=./mock/action_mock.go -package=mock
```
