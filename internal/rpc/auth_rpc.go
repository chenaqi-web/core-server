package rpc

import (
	"context"
	"errors"

	"backend/core-server/internal/application"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/rpc/authpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authApplication interface {
	SendEmailCode(ctx context.Context, email string, purpose domain.EmailCodePurpose) error
	Register(ctx context.Context, input application.RegisterInput) (*application.AuthUser, error)
	Login(ctx context.Context, username, password string) (*application.LoginResult, error)
	EmailLogin(ctx context.Context, email, password string) (*application.LoginResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*application.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	ResetPasswordByEmail(ctx context.Context, input application.ResetPasswordByEmailInput) error
	ValidateAccess(ctx context.Context, accessToken string) (*application.AccessIdentity, error)
}

type AuthRPC struct {
	authpb.UnimplementedAuthServiceServer
	service authApplication
}

func NewAuthRPC(service *application.AuthService) *AuthRPC {
	return newAuthRPC(service)
}

func newAuthRPC(service authApplication) *AuthRPC {
	return &AuthRPC{service: service}
}

func (r *AuthRPC) SendEmailCode(
	ctx context.Context,
	request *authpb.SendEmailCodeRequest,
) (*authpb.SendEmailCodeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	purpose, err := emailCodePurposeFromProto(request.GetPurpose())
	if err != nil {
		return nil, err
	}
	if err := r.service.SendEmailCode(ctx, request.GetEmail(), purpose); err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.SendEmailCodeResponse{Success: true}, nil
}

func (r *AuthRPC) Register(
	ctx context.Context,
	request *authpb.RegisterRequest,
) (*authpb.RegisterResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	user, err := r.service.Register(ctx, application.RegisterInput{
		Username:        request.GetUsername(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
		EmailCode:       request.GetEmailCode(),
		Phone:           request.GetPhone(),
		Avatar:          request.GetAvatar(),
		Sex:             request.GetSex(),
		Age:             request.GetAge(),
	})
	if err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.RegisterResponse{
		Success: true,
		User:    authUserToProto(user),
	}, nil
}

func (r *AuthRPC) Login(
	ctx context.Context,
	request *authpb.LoginRequest,
) (*authpb.LoginResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	result, err := r.service.Login(ctx, request.GetUsername(), request.GetPassword())
	if err != nil {
		return nil, authRPCError(err)
	}
	return loginResultToProto(result), nil
}

func (r *AuthRPC) EmailLogin(
	ctx context.Context,
	request *authpb.EmailLoginRequest,
) (*authpb.LoginResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	result, err := r.service.EmailLogin(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, authRPCError(err)
	}
	return loginResultToProto(result), nil
}

func (r *AuthRPC) RefreshToken(
	ctx context.Context,
	request *authpb.RefreshTokenRequest,
) (*authpb.RefreshTokenResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	tokens, err := r.service.RefreshToken(ctx, request.GetRefreshToken())
	if err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.RefreshTokenResponse{Tokens: authTokensToProto(tokens)}, nil
}

func (r *AuthRPC) Logout(
	ctx context.Context,
	request *authpb.LogoutRequest,
) (*authpb.LogoutResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := r.service.Logout(ctx, request.GetRefreshToken()); err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.LogoutResponse{Success: true}, nil
}

func (r *AuthRPC) ResetPasswordByEmail(
	ctx context.Context,
	request *authpb.ResetPasswordByEmailRequest,
) (*authpb.ResetPasswordByEmailResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := r.service.ResetPasswordByEmail(ctx, application.ResetPasswordByEmailInput{
		Email:           request.GetEmail(),
		EmailCode:       request.GetEmailCode(),
		NewPassword:     request.GetNewPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}); err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.ResetPasswordByEmailResponse{Success: true}, nil
}

func (r *AuthRPC) ValidateAccess(
	ctx context.Context,
	request *authpb.ValidateAccessRequest,
) (*authpb.ValidateAccessResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	identity, err := r.service.ValidateAccess(ctx, request.GetAccessToken())
	if err != nil {
		return nil, authRPCError(err)
	}
	return &authpb.ValidateAccessResponse{
		UserId:      identity.UserID,
		SessionId:   identity.SessionID,
		Role:        identity.Role,
		Status:      identity.Status,
		AuthVersion: identity.AuthVersion,
	}, nil
}

func emailCodePurposeFromProto(value authpb.EmailCodePurpose) (domain.EmailCodePurpose, error) {
	switch value {
	case authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_REGISTER:
		return domain.EmailCodePurposeRegister, nil
	case authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_RESET_PASSWORD:
		return domain.EmailCodePurposeResetPassword, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid email code purpose")
	}
}

func authUserToProto(user *application.AuthUser) *authpb.UserInfo {
	if user == nil {
		return nil
	}
	return &authpb.UserInfo{
		Id:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		Avatar:      user.Avatar,
		Sex:         user.Sex,
		Age:         uint32(user.Age),
		Role:        user.Role,
		Status:      user.Status,
		AuthVersion: user.AuthVersion,
	}
}

func loginResultToProto(result *application.LoginResult) *authpb.LoginResponse {
	if result == nil {
		return nil
	}
	return &authpb.LoginResponse{
		User:   authUserToProto(result.User),
		Tokens: authTokensToProto(result.Tokens),
	}
}

func authTokensToProto(tokens *application.AuthTokens) *authpb.TokenPair {
	if tokens == nil {
		return nil
	}
	return &authpb.TokenPair{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiresIn:  tokens.AccessExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
	}
}

func authRPCError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidAuthInput),
		errors.Is(err, application.ErrInvalidEmail),
		errors.Is(err, application.ErrInvalidEmailPurpose),
		errors.Is(err, application.ErrPasswordRule):
		return status.Error(codes.InvalidArgument, "invalid auth request")
	case errors.Is(err, application.ErrEmailCodeInvalid):
		return status.Error(codes.FailedPrecondition, "email code is invalid or expired")
	case errors.Is(err, application.ErrUsernameExists):
		return status.Error(codes.AlreadyExists, "username already exists")
	case errors.Is(err, application.ErrEmailExists):
		return status.Error(codes.AlreadyExists, "email already exists")
	case errors.Is(err, application.ErrEmailCodeCooldown),
		errors.Is(err, application.ErrEmailCodeHourlyLimit):
		return status.Error(codes.ResourceExhausted, "email sending rate limit exceeded")
	case errors.Is(err, application.ErrMailUnavailable):
		return status.Error(codes.Unavailable, "mail service unavailable")
	case errors.Is(err, application.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "username, email or password is incorrect")
	case errors.Is(err, application.ErrActiveSession):
		return status.Error(codes.AlreadyExists, "user already has an active session")
	case errors.Is(err, application.ErrUserDisabled):
		return status.Error(codes.PermissionDenied, "user is disabled")
	case errors.Is(err, application.ErrAccessInvalid):
		return status.Error(codes.Unauthenticated, "access token is invalid or expired")
	case errors.Is(err, application.ErrRefreshInvalid):
		return status.Error(codes.Unauthenticated, "refresh token is invalid or expired")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
