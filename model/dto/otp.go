package dto

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type VerifyOTPRequest struct {
    Email  string `json:"email" binding:"required,email"`
    OTP    string `json:"otp" binding:"required"`
}
