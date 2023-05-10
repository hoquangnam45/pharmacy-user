package data

import (
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type UserInfoRepo interface {
}

type userInfoRepo struct {
	data *Data
	log  log.Logger
}

func NewUserInfoRepo(data *Data, logger log.Logger) UserInfoRepo {
	return &userInfoRepo{
		data: data,
		log:  logger,
	}
}
