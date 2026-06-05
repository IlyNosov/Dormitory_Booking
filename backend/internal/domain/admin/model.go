package admin

import "time"

const SuperAdmin = "irnosov@edu.hse.ru"

type Admin struct {
	Email     string    `json:"email"`
	AddedBy   string    `json:"addedBy"`
	CreatedAt time.Time `json:"createdAt"`
}

func IsSuperAdmin(email string) bool {
	return email == SuperAdmin
}
