package service

import (
	"context"
	"time"

	"github.com/hellofresh/health-go/v5"
	v1 "github.com/hoquangnam45/pharmacy-user/api/user/v1"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type HealthCheckService struct {
	v1.UnimplementedHealthCheckServer
}

var healthCheck *health.Health

func NewHealthCheckService() *HealthCheckService {
	return &HealthCheckService{}
}

func (c *HealthCheckService) HealthCheck(ctx context.Context, req *emptypb.Empty) (*structpb.Struct, error) {
	return h.FlatMap3(
		h.LiftJ(healthCheck.Measure)(ctx),
		h.Lift(util.MarshalJson[health.Check]),
		h.Lift(util.UnmarshalJsonDeref(&map[string]any{})),
		h.Lift(structpb.NewStruct),
	).EvalWithContext(ctx)
}

func init() {
	h, err := health.New(health.WithComponent(health.Component{
		Name:    "pharmacy-auth",
		Version: "0.0.1-SNAPSHOT",
	}), health.WithChecks(health.Config{
		Name:      "ping",
		Timeout:   time.Second,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			return nil
		}},
	))
	if err != nil {
		panic(err)
	}
	healthCheck = h
}
