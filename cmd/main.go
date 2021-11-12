package main

import (
	"activity_log/api/constants"
	"activity_log/internal/chatter"
	datadao "activity_log/internal/dao/data_dao"
	schemadao "activity_log/internal/dao/schema_dao"
	"activity_log/internal/user_input"
	cli "activity_log/internal/user_input/service"
	"activity_log/internal/user_output"
	"time"
)

func main() {

	userListener := user_input.New(&cli.CLIListener{})
	userMessenger := &user_output.UserMessenger{}

	userSchemaDAO := schemadao.NewLocalSchemaDAO(constants.DEFAULT_SCHEMA_PATH)
	userDataDAO := datadao.NewDataDAO(constants.DEFAULT_DATA_PATH)

	chatterConfig := &chatter.ChatterConfig{
		ResponseWait:        time.Minute,
		MaxConfusionRetries: 3,
	}

	chatter := chatter.NewChatter(userListener, userMessenger, userSchemaDAO, userDataDAO, chatterConfig)

	chatter.Run()
}
