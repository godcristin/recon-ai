package biz

import (
	"errors"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewReconUseCase)

var (
	ErrEmptyFile       = errors.New("csv file is empty")
	ErrParseFailed     = errors.New("parse csv failed")
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidAction   = errors.New("invalid action")
	ErrDiffNotFound    = errors.New("diff not found")
	ErrDiffSettled     = errors.New("diff already settled")
)
