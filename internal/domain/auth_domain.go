package domain

import "context"

type EmailCodePurpose string

const (
	EmailCodePurposeRegister      EmailCodePurpose = "register"
	EmailCodePurposeResetPassword EmailCodePurpose = "reset_password"
)

func (p EmailCodePurpose) Valid() bool {
	return p == EmailCodePurposeRegister || p == EmailCodePurposeResetPassword
}

type MailSender interface {
	SendVerificationCode(ctx context.Context, recipient, code string, purpose EmailCodePurpose) error
}

type EmailCodeReservation struct {
	Code       string
	ID         string
	HourBucket string
}

type EmailCodeStore interface {
	ReserveCode(ctx context.Context, email string, purpose EmailCodePurpose) (*EmailCodeReservation, error)
	CancelCode(ctx context.Context, email string, purpose EmailCodePurpose, reservation *EmailCodeReservation) (bool, error)
	VerifyCode(ctx context.Context, email string, purpose EmailCodePurpose, code string) (bool, error)
}

type AuthSession struct {
	SessionID   string
	RefreshJTI  string
	AuthVersion uint64
}

type AuthSessionStore interface {
	CreateSession(ctx context.Context, userID uint64, session AuthSession) (bool, error)
	GetSession(ctx context.Context, userID uint64) (*AuthSession, error)
	ValidateRefreshJTI(ctx context.Context, userID uint64, sessionID, refreshJTI string) (bool, error)
	RotateRefreshJTI(ctx context.Context, userID uint64, sessionID, oldRefreshJTI, newRefreshJTI string) (bool, error)
	DeleteSessionIfMatch(ctx context.Context, userID uint64, sessionID, refreshJTI string) (bool, error)
	ClearUserSession(ctx context.Context, userID uint64) error
}
