// Code generated by entc, DO NOT EDIT.

package ent

import (
	"account/ent/schema"
	"account/ent/user"
	"account/ent/userinfo"
	"time"

	"github.com/google/uuid"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	userFields := schema.User{}.Fields()
	_ = userFields
	// userDescUsername is the schema descriptor for username field.
	userDescUsername := userFields[0].Descriptor()
	// user.UsernameValidator is a validator for the "username" field. It is called by the builders before save.
	user.UsernameValidator = userDescUsername.Validators[0].(func(string) error)
	// userDescPassword is the schema descriptor for password field.
	userDescPassword := userFields[1].Descriptor()
	// user.PasswordValidator is a validator for the "password" field. It is called by the builders before save.
	user.PasswordValidator = userDescPassword.Validators[0].(func(string) error)
	// userDescUUID is the schema descriptor for uuid field.
	userDescUUID := userFields[2].Descriptor()
	// user.DefaultUUID holds the default value on creation for the uuid field.
	user.DefaultUUID = userDescUUID.Default.(func() uuid.UUID)
	// userDescGmtCreate is the schema descriptor for gmt_create field.
	userDescGmtCreate := userFields[3].Descriptor()
	// user.DefaultGmtCreate holds the default value on creation for the gmt_create field.
	user.DefaultGmtCreate = userDescGmtCreate.Default.(func() time.Time)
	// userDescGmtModified is the schema descriptor for gmt_modified field.
	userDescGmtModified := userFields[4].Descriptor()
	// user.DefaultGmtModified holds the default value on creation for the gmt_modified field.
	user.DefaultGmtModified = userDescGmtModified.Default.(func() time.Time)
	userinfoFields := schema.UserInfo{}.Fields()
	_ = userinfoFields
	// userinfoDescUUID is the schema descriptor for uuid field.
	userinfoDescUUID := userinfoFields[0].Descriptor()
	// userinfo.DefaultUUID holds the default value on creation for the uuid field.
	userinfo.DefaultUUID = userinfoDescUUID.Default.(func() uuid.UUID)
	// userinfoDescGmtCreate is the schema descriptor for gmt_create field.
	userinfoDescGmtCreate := userinfoFields[1].Descriptor()
	// userinfo.DefaultGmtCreate holds the default value on creation for the gmt_create field.
	userinfo.DefaultGmtCreate = userinfoDescGmtCreate.Default.(func() time.Time)
	// userinfoDescGmtModified is the schema descriptor for gmt_modified field.
	userinfoDescGmtModified := userinfoFields[2].Descriptor()
	// userinfo.DefaultGmtModified holds the default value on creation for the gmt_modified field.
	userinfo.DefaultGmtModified = userinfoDescGmtModified.Default.(func() time.Time)
	// userinfoDescNickname is the schema descriptor for nickname field.
	userinfoDescNickname := userinfoFields[3].Descriptor()
	// userinfo.NicknameValidator is a validator for the "nickname" field. It is called by the builders before save.
	userinfo.NicknameValidator = userinfoDescNickname.Validators[0].(func(string) error)
	// userinfoDescAvatar is the schema descriptor for avatar field.
	userinfoDescAvatar := userinfoFields[4].Descriptor()
	// userinfo.DefaultAvatar holds the default value on creation for the avatar field.
	userinfo.DefaultAvatar = userinfoDescAvatar.Default.(string)
}
