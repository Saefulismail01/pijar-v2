package model

import "time"

type OTP struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email"`
    Code      string    `json:"code"`
    ExpiresAt time.Time `json:"expires_at"`
    Attempts  int       `json:"attempts"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// OTPRepository interface untuk repository OTP
type OTPRepository interface {
    SaveOTP(otp *OTP) error
    GetOTPByCode(code string) (*OTP, error)
    DeleteExpiredOTPs() error
}
