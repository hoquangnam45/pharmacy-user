package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type clientRepo struct {
	data *Data
	log  log.Logger
}

func NewClientRepo(data *Data, logger log.Logger) biz.ClientRepo {
	return &clientRepo{
		data: data,
		log:  logger,
	}
}

func (r *clientRepo) FindByClientID(clientId string) (*biz.Client, error) {
	data := &biz.Client{}
	err := r.data.Where(&biz.Client{ClientId: clientId}).Take(data).Error
	return data, err
}
