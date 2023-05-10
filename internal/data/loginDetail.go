package data

import (
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type loginDetailRepo struct {
	data *Data
	log  log.Logger
}

func NewLoginDetailRepo(data *Data, logger log.Logger) biz.LoginDetailRepo {
	return &loginDetailRepo{
		data: data,
		log:  logger,
	}
}

func (r *loginDetailRepo) Save(g *biz.LoginDetail) (*biz.LoginDetail, error) {
	if err := r.data.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (r *loginDetailRepo) FindByID(id uuid.UUID) (*biz.LoginDetail, error) {
	loginDetail := &biz.LoginDetail{}
	err := r.data.Where(&biz.LoginDetail{Id: id}).Take(loginDetail).Error
	if err != nil {
		return nil, err
	}
	return loginDetail, nil
}

func (r *loginDetailRepo) FindByUserId(userId string) (*biz.LoginDetail, error) {
	loginDetail := &biz.LoginDetail{}
	err := r.data.Where(&biz.LoginDetail{UserId: userId}).Take(loginDetail).Error
	if err != nil {
		return nil, err
	}
	return loginDetail, nil
}
