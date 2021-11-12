package dao

import "activity_log/api/constructs"

type UserSchemaDAO interface {
	Load() (*constructs.UserSchema, error)
	Dump(schema *constructs.UserSchema, force bool) error
	Init() (*constructs.UserSchema, error)
}

type UserDataDAO interface {
	Append(data *constructs.UserData) error
}
