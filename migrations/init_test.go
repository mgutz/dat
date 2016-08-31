package migrations

func init() {
	// setup testing
	userOptions := NewDBOptions()
	superOptions := NewDBOptions()
	superOptions.Password = "password"
	Init(userOptions, superOptions, "GO", "")
}
