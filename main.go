package main

import (
	"github.com/franzwilhelm/uio-exam-helper/cmd"
	_ "github.com/franzwilhelm/uio-exam-helper/db"
	"github.com/franzwilhelm/uio-exam-helper/db/migration"
)

func main() {
	migration.MigrateAll()
	cmd.Execute()
}
