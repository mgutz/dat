package migrations

func init() {
	// setup testing
	userOptions := NewDBOptions()
	superOptions := NewDBOptions()

	// dbname=dat_test user=dat password=!test host=localhost sslmode=disable
	userOptions.DBName = "dat_test"
	userOptions.Password = "!test"
	userOptions.User = "dat"
	userOptions.SSLMode = false

	superOptions.DBName = "postgres"
	superOptions.Password = "password"
	superOptions.User = "postgres"
	superOptions.SSLMode = false

	Init(userOptions, superOptions)
}
