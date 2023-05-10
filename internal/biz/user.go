package biz

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-common-go/helper/db"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
	"github.com/hoquangnam45/pharmacy-common-go/util/request"
	"gorm.io/gorm"
)

type UserInfo struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	UserId    string
	Password  string
	Activated bool
}

type UserInfoRepo interface {
	Save(*UserInfo) (*UserInfo, error)
	FindByID(uuid.UUID) (*UserInfo, error)
	FindById(userId string) (*UserInfo, error)
}

type RefreshTokenRepo interface {
	Save(*RefreshToken) (*RefreshToken, error)
	DeleteById(id string) error
	FindById(id string) (*RefreshToken, error)
}

type TransactionManager interface {
	Run(func() error) error
}

type ClientRepo interface {
	FindByClientID(clientId string) (*Client, error)
}

func Query[T any](txManager TransactionManager, queryFn func() (T, error)) (T, error) {
	var ret T
	err := txManager.Run(func() error {
		retI, err := queryFn()
		if err == nil {
			ret = retI
		}
		return err
	})
	return ret, err
}

type TrustedTpInfoFetcher func(accessToken string) (*UserInfo, error)

type LoginDetailUsecase struct {
	repo             LoginDetailRepo
	refreshTokenRepo RefreshTokenRepo
	clientRepo       ClientRepo
	userInfoClient   UserInfoClient
	TransactionManager
	log log.Logger
}

func NewLoginDetailUseCase(userInfoClient UserInfoClient, repo LoginDetailRepo, clientRepo ClientRepo, refreshTokenRepo RefreshTokenRepo, logger log.Logger, transactionManager TransactionManager) *LoginDetailUsecase {
	return &LoginDetailUsecase{
		repo:               repo,
		refreshTokenRepo:   refreshTokenRepo,
		clientRepo:         clientRepo,
		log:                logger,
		TransactionManager: transactionManager,
		userInfoClient:     userInfoClient,
	}
}

func (s *LoginDetailUsecase) GenerateAccessToken(refreshToken *RefreshToken) (string, error) {
	claims := AuthClaims{}
	client := refreshToken.Client
	claims.ExpiresAt = jwt.NewNumericDate(refreshToken.IssuedAt.Add(client.AccessTokenTtl.ToDuration()))
	claims.IssuedAt = jwt.NewNumericDate(refreshToken.IssuedAt)
	claims.ID = uuid.New().String()
	claims.Issuer = client.Issuer
	claims.Subject = refreshToken.Subject
	signingKey := client.SigningKey
	signingMethod := jwt.GetSigningMethod(client.SigningMethod)
	token := jwt.NewWithClaims(signingMethod, claims)
	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(signingKey))
	if err == nil {
		return token.SignedString(privateKey)
	} else {
		return "", err
	}
}

func (s *LoginDetailUsecase) Activate(id string) error {
	return h.FlatMap2(
		h.Lift(uuid.Parse)(id),
		h.Lift(s.repo.FindByID),
		h.LiftE(func(loginDetail *LoginDetail) error {
			if !loginDetail.Activated {
				loginDetail.Activated = true
				return h.Lift(s.repo.Save)(loginDetail).Error()
			}
			return nil
		}),
	).Error()
}

func (s *LoginDetailUsecase) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteById(refreshToken)
}

func (s *LoginDetailUsecase) FindClient(clientId string) (*Client, error) {
	return h.Lift(s.clientRepo.FindByClientID)(clientId).EvalWithHandlerE(func(err error) error {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidClientId
		}
		return nil
	})
}

func (s *LoginDetailUsecase) Login(grantRequest *GrantRequest) (*Authentication, error) {
	switch grantRequest.GrantType.Normalize() {
	case grantType.Password:
		return s.passwordAuthenticated(grantRequest)
	case grantType.RefreshToken:
		return s.refreshTokenAuthenticated(grantRequest)
	case grantType.TrustedTp:
		switch grantRequest.Provider.Normalize() {
		case oauthProviderType.Facebook:
			return s.trustedTpAuthenticated(FetchFbUserInfo, *grantRequest.AccessToken, grantRequest.ClientId)
		case oauthProviderType.Google:
			return s.trustedTpAuthenticated(FetchGoogleUserInfo, *grantRequest.AccessToken, grantRequest.ClientId)
		default:
			return nil, fmt.Errorf("%w %s", ErrNotSupportOauthProvider, *grantRequest.Provider)
		}
	default:
		return nil, fmt.Errorf("%w %s", ErrNotSupportGrantType, grantRequest.GrantType)
	}
}

func (s *LoginDetailUsecase) Register(registerRequest *GrantRequest) (*Authentication, error) {
	return h.FlatMap2(
		h.FactoryM(func() (*UserInfo, error) {
			return s.userInfoClient.CreateUserInfo(registerRequest.Username, registerRequest.Email, registerRequest.PhoneNumber)
		}),
		h.Lift(func(userInfo *UserInfo) (*LoginDetail, error) {
			return h.FlatMap(
				h.Lift(util.HashPassword)(*registerRequest.Password),
				h.Lift(func(pass string) (*LoginDetail, error) {
					return s.repo.Save(&LoginDetail{
						UserId:    userInfo.Id,
						Password:  pass,
						Activated: false,
					})
				})).Eval()
		}),
		h.LiftJ(func(loginDetail *LoginDetail) *Authentication {
			return &Authentication{
				Subject:       loginDetail.UserId,
				Authenticated: true,
				Credential:    loginDetail.Password,
				GrantType:     registerRequest.GrantType,
				ClientId:      registerRequest.ClientId,
			}
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		groupErr := &util.GroupError{}
		if !errors.As(err, groupErr) || !errors.Is(groupErr.Group, ErrUserInfoClientGroup) || !errors.Is(groupErr.Cause, ErrResourceAlreadyExists) {
			s.userInfoClient.RemoveUserInfo(registerRequest.Username, registerRequest.Email, registerRequest.Password)
		} else {
			// Credential should already exist here
			return ErrCredentialAlreadyExist
		}
		if db.IsDuplicatedError(err) || errors.As(err, &requestErr) && requestErr.StatusCode == 409 {
			// This should not be happening, it should return err from user info client first, please check the db for inconsistency again
			s.log.Error("inconsistency between auth service and user info service of user[email=%s, username=%s]", registerRequest.Email, registerRequest.Username)
			return ErrCredentialAlreadyExist
		}
		return nil
	})
}

func (s *LoginDetailUsecase) GrantAccess(authentication *Authentication) (*GrantAccess, error) {
	if authentication == nil || !authentication.Authenticated {
		return nil, ErrUnauthorizedAccess
	}

	client, err := h.Lift(s.clientRepo.FindByClientID)(authentication.ClientId).EvalWithHandlerE(func(err error) error {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidClientId
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else if !client.Active {
		return nil, ErrUnauthorizedAccess
	}
	switch authentication.GrantType.Normalize() {
	case grantType.Password:
		fallthrough
	case grantType.TrustedTp:
		return s.grantAccessCommon(authentication, client)
	case grantType.RefreshToken:
		return s.grantAccessByRefreshToken(authentication, client)
	default:
		return nil, ErrUnauthorizedAccess
	}
}

func (s *LoginDetailUsecase) grantAccessByRefreshToken(authentication *Authentication, client *Client) (*GrantAccess, error) {
	var grantAccess *GrantAccess = nil
	err := s.Run(func() error {
		refreshToken := authentication.Credential.(*RefreshToken)
		now := time.Now()
		if err := s.refreshTokenRepo.DeleteById(refreshToken.Id); err != nil {
			return err
		}
		newRefreshToken, err := s.refreshTokenRepo.Save(&RefreshToken{
			Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
			IssuedAt:  now,
			ExpiredAt: now.Add(client.RefreshTokenTtl.ToDuration()),
			ClientId:  client.Id,
			Client:    client,
			Subject:   authentication.Subject,
		})
		if err != nil {
			return err
		}
		accessToken, err := s.GenerateAccessToken(newRefreshToken)
		if err != nil {
			return err
		}
		grantAccess = &GrantAccess{
			RefreshToken: newRefreshToken.Id,
			AccessToken:  accessToken,
			Subject:      newRefreshToken.Subject,
			IssuedAt:     newRefreshToken.IssuedAt,
			ExpiredAt:    newRefreshToken.ExpiredAt,
			ExpiredIn:    client.AccessTokenTtl.ToDuration(),
			ClientId:     client.ClientId,
		}
		return nil
	})
	return grantAccess, err
}

func (s *LoginDetailUsecase) grantAccessCommon(authentication *Authentication, client *Client) (*GrantAccess, error) {
	now := time.Now()
	newRefreshToken := &RefreshToken{
		Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
		IssuedAt:  now,
		ExpiredAt: now.Add(client.RefreshTokenTtl.ToDuration()),
		Client:    client,
		ClientId:  client.Id,
		Subject:   authentication.Subject,
	}
	accessToken, err := s.GenerateAccessToken(newRefreshToken)
	if err != nil {
		return nil, err
	}
	grantAccess := &GrantAccess{
		RefreshToken: newRefreshToken.Id,
		AccessToken:  accessToken,
		Subject:      newRefreshToken.Subject,
		IssuedAt:     newRefreshToken.IssuedAt,
		ExpiredAt:    newRefreshToken.ExpiredAt,
		ExpiredIn:    client.AccessTokenTtl.ToDuration(),
		ClientId:     client.ClientId,
	}
	if _, err = s.refreshTokenRepo.Save(newRefreshToken); err != nil {
		return nil, err
	}
	return grantAccess, nil
}

func (s *LoginDetailUsecase) trustedTpAuthenticated(fetcher TrustedTpInfoFetcher, accessToken string, clientId string) (*Authentication, error) {
	return h.FlatMap2(
		h.Lift(fetcher)(accessToken),
		h.Lift(func(userInfo *UserInfo) (*UserInfo, error) {
			if userInfo, err := s.userInfoClient.FetchUserInfo(nil, &userInfo.Email, nil); errors.Is(err, ErrResourceAlreadyExists) {
				return s.userInfoClient.CreateUserInfo(nil, &userInfo.Email, nil)
			} else {
				return userInfo, nil
			}
		}),
		h.Lift(func(userInfo *UserInfo) (*Authentication, error) {
			return &Authentication{
				Subject:       userInfo.Id,
				Authenticated: true,
				Credential:    accessToken,
				GrantType:     grantType.TrustedTp,
				ClientId:      clientId,
			}, nil
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		if errors.As(err, &requestErr) || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredential
		}
		return nil
	})
}

func (s *LoginDetailUsecase) passwordAuthenticated(loginRequest *GrantRequest) (*Authentication, error) {
	return h.FlatMap(
		h.FactoryM(func() (*UserInfo, error) {
			return s.userInfoClient.FetchUserInfo(loginRequest.Username, loginRequest.Email, loginRequest.PhoneNumber)
		}),
		h.Lift(func(userInfo *UserInfo) (*Authentication, error) {
			loginDetail, err := s.repo.FindByUserId(userInfo.Id)
			if err != nil {
				return nil, err
			}
			if !util.ComparePassword(*loginRequest.Password, loginDetail.Password) {
				return nil, ErrInvalidCredential
			}
			return &Authentication{
				Subject:       userInfo.Id,
				Authenticated: true,
				Credential:    loginDetail.Password,
				GrantType:     loginRequest.GrantType,
				ClientId:      loginRequest.ClientId,
			}, nil
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		if errors.As(err, &requestErr) || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredential
		}
		return nil
	})
}

func (s *LoginDetailUsecase) refreshTokenAuthenticated(grantRequest *GrantRequest) (*Authentication, error) {
	var auth *Authentication = nil
	err := s.Run(func() error {
		refreshToken, err := s.refreshTokenRepo.FindById(*grantRequest.RefreshToken)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorizedAccess
		} else if err != nil {
			return err
		}
		if time.Now().After(refreshToken.ExpiredAt) {
			s.refreshTokenRepo.DeleteById(refreshToken.Id)
			return ErrUnauthorizedAccess
		}
		authI, err := h.FlatMap2(
			h.Lift(base64.StdEncoding.DecodeString)(refreshToken.ProtectedTicket),
			h.Lift(util.UnmarshalJson(&UserInfo{})),
			h.Lift(func(userInfo *UserInfo) (*Authentication, error) {
				return &Authentication{
					Authenticated: true,
					Subject:       refreshToken.Subject,
					Credential:    refreshToken,
					GrantType:     grantRequest.GrantType,
					ClientId:      grantRequest.ClientId,
				}, nil
			}),
		).Eval()
		if err != nil {
			return err
		}
		auth = authI
		return nil
	})
	return auth, err
}

// TODO: Implement this method
func (s *LoginDetailUsecase) CheckPermission() error {
	return nil
}
