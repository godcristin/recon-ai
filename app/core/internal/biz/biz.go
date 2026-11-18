package biz

import (
	"errors"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewReconUseCase)

// 领域错误：由 service 层翻译成对外错误码
var (
	ErrEmptyFile = errors.New("csv file is empty")
)
