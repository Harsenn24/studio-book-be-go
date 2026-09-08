package UserRepository

type User struct {
	ID              uint64 `json:"id"`                // BIGINT UNSIGNED biasanya dipetakan ke uint64 di Go
	UUID            string `json:"uuid"`              // VARCHAR -> string
	Name            string `json:"name"`              // VARCHAR -> string
	Email           string `json:"email"`             // VARCHAR -> string
	EmailVerifiedAt *int64 `json:"email_verified_at"` // Karena "Allow Null" dicentang, kita wajib pakai pointer (*int64) agar bisa bernilai nil
	Password        string `json:"-"`                // Disembunyikan dari JSON response
	Role            string `json:"role"`              // VARCHAR -> string
	CreatedAt       int64  `json:"created_at"`        // BIGINT UNSIGNED untuk Unix timestamp -> int64 atau uint64
	UpdatedAt       int64  `json:"updated_at"`        // BIGINT UNSIGNED untuk Unix timestamp -> int64 atau uint64
}

func (User) TableName() string {
	return "users"
}
