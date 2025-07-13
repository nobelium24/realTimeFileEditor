package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Media struct {
	Public_ID  string `gorm:"type:text(1024)" json:"public_id"`
	Secure_URL string `gorm:"type:text(1024)" json:"secure_url"`
}

type MediaList []Media

func (i *Media) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	return json.Marshal(i)
}

func (i *Media) Scan(value any) error {
	if value == nil {
		*i = Media{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal Images value: %v", value)
	}
	return json.Unmarshal(bytes, i)
}

func (m *MediaList) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *MediaList) Scan(value any) error {
	if value == nil {
		*m = MediaList{}
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("error unmarshaling Images slice value: %v", value)
	}

	return json.Unmarshal(bytes, m)
}
