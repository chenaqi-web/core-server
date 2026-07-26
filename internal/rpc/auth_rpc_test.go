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
	sendEmail     string
	sendPurpose   domain.EmailCodePurpose
	sendErr       error
	registerInput application.RegisterInput
	registerUser  *application.AuthUser
	registerErr   error
	resetInput    application.ResetPasswordByEmailInput
	resetErr      error
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

func (s *stubAuthApplication) ResetPasswordByEmail(
	_ context.Context,
	input application.ResetPasswordByEmailInput,
) error {
	s.resetInput = input
	return s.resetErr
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
