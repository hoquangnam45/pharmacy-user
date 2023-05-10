package biz

import (
	stderrors "errors"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/hoquangnam45/pharmacy-common-go/util"
)

// Standard errors
var ErrResourceAlreadyExists = stderrors.New("resource already exists")
var ErrResourceNotFound = stderrors.New("resource not found")

// Login detail error group
var ErrLoginDetailErrorGroup = stderrors.New("login detail error group")
var (
	ErrNotSupportOauthProvider = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "not supported openid provider", "not supported openid provider"))
	ErrNotSupportGrantType     = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "not supported grant type", "not supported grant type"))
	ErrInvalidTpAccessToken    = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "invalid trusted third party access token", "invalid trusted third party access token"))
	ErrInvalidCredential       = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "invalid credential", "invalid credential"))
	ErrCredentialAlreadyExist  = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(409, "crendential already exist", "crendential already exist"))
	ErrInvalidClientId         = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "invalid client id", "invalid client id"))
	ErrUnauthorizedAccess      = util.NewGroupError(ErrLoginDetailErrorGroup, errors.New(401, "unauthorized access", "unauthorized access"))
)

// User info client error group
var ErrUserInfoClientGroup = stderrors.New("user info client group")
