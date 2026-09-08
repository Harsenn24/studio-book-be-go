package EmailLoggerRepository

import "encoding/json"

type EmailLogger struct {
	ID        uint64           `json:"id"`      // BIGINT UNSIGNED biasanya dipetakan ke uint64 di Go
	UserID    uint64           `json:"user_id"` // BIGINT UNSIGNED biasanya dipetakan ke uint64 di Go            string `json:"uuid"`              // VARCHAR -> string
	Email     string           `json:"email"`   // VARCHAR -> string
	Type      string           `json:"type"`    // Karena "Allow Null" dicentang, kita wajib pakai pointer (*int64) agar bisa bernilai nil
	MetaData  *json.RawMessage `json:"meta_data"`
	CreatedAt int64            `json:"created_at"` // BIGINT UNSIGNED untuk Unix timestamp -> int64 atau uint64
	UpdatedAt int64            `json:"updated_at"` // BIGINT UNSIGNED untuk Unix timestamp -> int64 atau uint64
}

func (EmailLogger) TableName() string {
	return "email_loggers"
}
