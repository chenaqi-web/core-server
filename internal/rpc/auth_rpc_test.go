package rpc

import (
	"context"
	"errors"
	"testing"

	"backend/core-server/internal/application"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/rpc/authpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubAuthApplication struct {
	sendEmail       string
	sendPurpose     domain.EmailCodePurpose
	sendErr         error
	registerInput   application.RegisterInput
	registerUser    *application.AuthUser
	registerErr     error
	loginIdentifier string
	loginPassword   string
	loginResult     *application.LoginResult
	loginErr        error
	refreshToken    string
	refreshResult   *application.AuthTokens
	refreshErr      error
	logoutToken     string
	logoutErr       error
	resetInput      application.ResetPasswordByEmailInput
	resetErr        error
	accessToken     string
	accessIdentity  *application.AccessIdentity
	accessErr       error
}

func (s *stubAuthApplication) SendEmailCode(
	_ context.Context,
	email string,
	purpose domain.EmailCodePurpose,
) error {
	s.sendEmail = email
	s.sendPurpose = purpose
	return s.sendErr
}

func (s *stubAuthApplication) Register(
	_ context.Context,
	input application.RegisterInput,
) (*application.AuthUser, error) {
	s.registerInput = input
	return s.registerUser, s.registerErr
}

func (s *stubAuthApplication) Login(
	_ context.Context,
	username, password string,
) (*application.LoginResult, error) {
	s.loginIdentifier = username
	s.loginPassword = password
	return s.loginResult, s.loginErr
}

func (s *stubAuthApplication) EmailLogin(
	_ context.Context,
	email, password string,
) (*application.LoginResult, error) {
	s.loginIdentifier = email
	s.loginPassword = password
	return s.loginResult, s.loginErr
}

func (s *stubAuthApplication) RefreshToken(_ context.Context, refreshToken string) (*application.AuthTokens, error) {
	s.refreshToken = refreshToken
	return s.refreshResult, s.refreshErr
}

func (s *stubAuthApplication) Logout(_ context.Context, refreshToken string) error {
	s.logoutToken = refreshToken
	return s.logoutErr
}

func (s *stubAuthApplication) ResetPasswordByEmail(
	_ context.Context,
	input application.ResetPasswordByEmailInput,
) error {
	s.resetInput = input
	return s.resetErr
}

func (s *stubAuthApplication) ValidateAccess(_ context.Context, accessToken string) (*application.AccessIdentity, error) {
	s.accessToken = accessToken
	return s.accessIdentity, s.accessErr
}

func TestAuthRPCSendEmailCodeConvertsPurpose(t *testing.T) {
	service := &stubAuthApplication{}
	rpc := newAuthRPC(service)
	response, err := rpc.SendEmailCode(context.Background(), &authpb.SendEmailCodeRequest{
		Email:   "user@qq.com",
		Purpose: authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_RESET_PASSWORD,
	})
	if err != nil {
		t.Fatalf("SendEmailCode() error = %v", err)
	}
	if !response.GetSuccess() || service.sendEmail != "user@qq.com" || service.sendPurpose != domain.EmailCodePurposeResetPassword {
		t.Fatalf("response = %+v, email = %q, purpose = %q", response, service.sendEmail, service.sendPurpose)
	}
}

func TestAuthRPCSendEmailCodeRejectsUnknownPurpose(t *testing.T) {
	service := &stubAuthApplication{}
	_, err := newAuthRPC(service).SendEmailCode(context.Background(), &authpb.SendEmailCodeRequest{
		Email:   "user@qq.com",
		Purpose: authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_UNSPECIFIED,
	})
	if status.Code(err) != codes.InvalidArgument || service.sendEmail != "" {
		t.Fatalf("SendEmailCode() error = %v, service email = %q", err, service.sendEmail)
	}
}

func TestAuthRPCRegisterConvertsRequestAndSafeUserResponse(t *testing.T) {
	service := &stubAuthApplication{registerUser: &application.AuthUser{
		ID: 9, Username: "user", Email: "user@qq.com", Phone: "", Avatar: "", Sex: "unknown",
		Age: 18, Role: "user", Status: "active", AuthVersion: 1,
	}}
	response, err := newAuthRPC(service).Register(context.Background(), &authpb.RegisterRequest{
		Username: "user", Email: "user@qq.com", Password: "abc12345", ConfirmPassword: "abc12345",
		EmailCode: "123456", Phone: "", Avatar: "", Sex: "unknown", Age: 18,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if service.registerInput.Password != "abc12345" || service.registerInput.EmailCode != "123456" {
		t.Fatalf("register input = %+v", service.registerInput)
	}
	if !response.GetSuccess() || response.GetUser().GetId() != 9 || response.GetUser().GetRole() != "user" {
		t.Fatalf("Register() response = %+v", response)
	}
}

func TestAuthRPCLoginAndEmailLoginReturnInternalTokenPair(t *testing.T) {
	methods := []struct {
		name string
		call func(*AuthRPC) (*authpb.LoginResponse, error)
		want string
	}{
		{name: "username", call: func(rpc *AuthRPC) (*authpb.LoginResponse, error) {
			return rpc.Login(context.Background(), &authpb.LoginRequest{Username: "user", Password: "abc12345"})
		}, want: "user"},
		{name: "email", call: func(rpc *AuthRPC) (*authpb.LoginResponse, error) {
			return rpc.EmailLogin(context.Background(), &authpb.EmailLoginRequest{Email: "user@qq.com", Password: "abc12345"})
		}, want: "user@qq.com"},
	}
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			service := &stubAuthApplication{loginResult: &application.LoginResult{
				User: &application.AuthUser{ID: 7, Username: "user", Role: "user", Status: "active", AuthVersion: 1},
				Tokens: &application.AuthTokens{
					AccessToken: "access", RefreshToken: "refresh", AccessExpiresIn: 1200, RefreshExpiresIn: 604800,
				},
			}}
			response, err := method.call(newAuthRPC(service))
			if err != nil {
				t.Fatalf("login RPC error = %v", err)
			}
			if service.loginIdentifier != method.want || service.loginPassword != "abc12345" ||
				response.GetTokens().GetAccessToken() != "access" || response.GetTokens().GetRefreshToken() != "refresh" {
				t.Fatalf("identifier = %q, response = %+v", service.loginIdentifier, response)
			}
		})
	}
}

func TestAuthRPCRefreshLogoutAndValidateAccessConvertRequests(t *testing.T) {
	service := &stubAuthApplication{
		refreshResult: &application.AuthTokens{AccessToken: "new-access", RefreshToken: "new-refresh"},
		accessIdentity: &application.AccessIdentity{
			UserID: 7, SessionID: "session", Role: "admin", Status: "active", AuthVersion: 3,
		},
	}
	rpc := newAuthRPC(service)
	refreshResponse, err := rpc.RefreshToken(context.Background(), &authpb.RefreshTokenRequest{RefreshToken: "old-refresh"})
	if err != nil || refreshResponse.GetTokens().GetRefreshToken() != "new-refresh" || service.refreshToken != "old-refresh" {
		t.Fatalf("RefreshToken() response = %+v, error = %v", refreshResponse, err)
	}
	logoutResponse, err := rpc.Logout(context.Background(), &authpb.LogoutRequest{RefreshToken: "new-refresh"})
	if err != nil || !logoutResponse.GetSuccess() || service.logoutToken != "new-refresh" {
		t.Fatalf("Logout() response = %+v, error = %v", logoutResponse, err)
	}
	accessResponse, err := rpc.ValidateAccess(context.Background(), &authpb.ValidateAccessRequest{AccessToken: "access"})
	if err != nil || accessResponse.GetUserId() != 7 || accessResponse.GetRole() != "admin" || service.accessToken != "access" {
		t.Fatalf("ValidateAccess() response = %+v, error = %v", accessResponse, err)
	}
}

func TestAuthRPCResetPasswordConvertsRequest(t *testing.T) {
	service := &stubAuthApplication{}
	response, err := newAuthRPC(service).ResetPasswordByEmail(context.Background(), &authpb.ResetPasswordByEmailRequest{
		Email: "user@qq.com", EmailCode: "123456", NewPassword: "newabc123", ConfirmPassword: "newabc123",
	})
	if err != nil {
		t.Fatalf("ResetPasswordByEmail() error = %v", err)
	}
	if !response.GetSuccess() || service.resetInput.NewPassword != "newabc123" || service.resetInput.EmailCode != "123456" {
		t.Fatalf("response = %+v, input = %+v", response, service.resetInput)
	}
}

func TestAuthRPCErrorMappingDoesNotLeakInternalErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "invalid", err: application.ErrPasswordRule, code: codes.InvalidArgument},
		{name: "code", err: application.ErrEmailCodeInvalid, code: codes.FailedPrecondition},
		{name: "exists", err: application.ErrEmailExists, code: codes.AlreadyExists},
		{name: "rate", err: application.ErrEmailCodeCooldown, code: codes.ResourceExhausted},
		{name: "mail", err: application.ErrMailUnavailable, code: codes.Unavailable},
		{name: "credentials", err: application.ErrInvalidCredentials, code: codes.Unauthenticated},
		{name: "session", err: application.ErrActiveSession, code: codes.AlreadyExists},
		{name: "disabled", err: application.ErrUserDisabled, code: codes.PermissionDenied},
		{name: "access", err: application.ErrAccessInvalid, code: codes.Unauthenticated},
		{name: "refresh", err: application.ErrRefreshInvalid, code: codes.Unauthenticated},
		{name: "internal", err: errors.New("mysql password secret"), code: codes.Internal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped := authRPCError(test.err)
			if status.Code(mapped) != test.code {
				t.Fatalf("authRPCError() code = %v, want %v", status.Code(mapped), test.code)
			}
			if test.name == "internal" && status.Convert(mapped).Message() != "internal error" {
				t.Fatalf("internal error leaked: %v", mapped)
			}
		})
	}
}
