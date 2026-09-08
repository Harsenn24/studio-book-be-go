package UserLoggerRepository

type UserLogger struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"` // INTEGER, Primary Key, Auto Increment tidak wajib pakai pointer
	UserID    uint    `gorm:"column:user_id;not null" json:"user_id"` // INTEGER, Allow Null: false
	Action    *string `gorm:"column:action" json:"action"`            // STRING, Allow Null: true -> pointer
	Success   *bool   `gorm:"column:success" json:"success"`          // BOOLEAN, Allow Null: true -> pointer
	IPAddress *string `gorm:"column:ip_address" json:"ip_address"`    // STRING, Allow Null: true -> pointer
	DeviceID  *string `gorm:"column:device_id" json:"device_id"`      // STRING, Allow Null: true -> pointer
	CreatedAt int64 `gorm:"column:created_at" json:"created_at"`    // BIGINT UNSIGNED, Allow Null: true -> pointer
	UpdatedAt int64 `gorm:"column:updated_at" json:"updated_at"`    // BIGINT UNSIGNED, Allow Null: true -> pointer
}

func (UserLogger) TableName() string {
	return "user_logs"
}