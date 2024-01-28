package logic

import (
	"context"

	"mooon-gateway/internal/svc"
	"mooon-gateway/pb/mooon_auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthenticateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthenticateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticateLogic {
	return &AuthenticateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthenticateLogic) Authenticate(in *mooon_auth.AuthReq) (*mooon_auth.AuthResp, error) {
	// todo: add your logic here and delete this line

	return &mooon_auth.AuthResp{}, nil
}
