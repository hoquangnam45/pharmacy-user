package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"

	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type refreshTokenRepo struct {
	data *Data
	log  log.Logger
}

func NewRefreshTokenRepo(data *Data, logger log.Logger) biz.RefreshTokenRepo {
	return &refreshTokenRepo{
		data: data,
		log:  logger,
	}
}

func (r *refreshTokenRepo) Save(g *biz.RefreshToken) (*biz.RefreshToken, error) {
	if err := r.data.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (r *refreshTokenRepo) DeleteById(id string) error {
	return r.data.Delete(&biz.RefreshToken{Id: id}).Error
}

func (r *refreshTokenRepo) FindById(id string) (*biz.RefreshToken, error) {
	data := &biz.RefreshToken{Id: id}
	err := r.data.Take(data).Error
	return data, err
}
