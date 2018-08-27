package migration

import (
	"github.com/franzwilhelm/uio-exam-helper/db"
	"github.com/franzwilhelm/uio-exam-helper/db/model"
)

func MigrateAll() error {
	return db.Default.AutoMigrate(
		&model.Resource{},
		&model.Subject{},
	).Error
}
