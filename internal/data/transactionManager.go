package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
	"gorm.io/gorm"
)

type transactionManager struct {
	data *Data
	log  log.Logger
}

func NewTransactionManager(data *Data, logger log.Logger) biz.TransactionManager {
	return &transactionManager{
		data: data,
		log:  logger,
	}
}

func (s *transactionManager) Run(f func() error) error {
	return s.data.Transaction(func(*gorm.DB) error {
		return f()
	})
}
