package auth

// LoginRequest digunakan untuk menangkap data JSON input dari client saat login
type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"` // binding:"required" artinya wajib diisi, dan harus format email
	Password  string `json:"password" binding:"required"`
	IPAddress string `json:"-"`
	DeviceID  string `json:"-"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type CheckUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type CheckUserResponse struct {
	Message string `json:"message"`
}

type VerifyEmailRequest struct {
	TokenVerify string `json:"token_verify" binding:"required"`
	Role        string `json:"role" binding:"required"`
}

type VerifyEmailResponse struct {
	Token string `json:"token"`
}
