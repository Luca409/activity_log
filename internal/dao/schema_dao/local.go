package schemadao

import (
	"activity_log/api/constants"
	"activity_log/api/constructs"
	"activity_log/internal/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type LocalSchemaDAO struct {
	path string
}

func NewLocalSchemaDAO(path string) *LocalSchemaDAO {
	return &LocalSchemaDAO{
		path: path,
	}
}

func (lsd *LocalSchemaDAO) Load() (*constructs.UserSchema, error) {
	us, err := loadUserSchema(lsd.path)
	if err != nil {
		return nil, fmt.Errorf("loadUserSchema(%s) returns err: %w", lsd.path, err)
	}
	return us, nil
}

func (lsd *LocalSchemaDAO) Dump(schema *constructs.UserSchema, force bool) error {
	return dumpUserSchema(lsd.path, schema)
}

func (lsd *LocalSchemaDAO) Init() (*constructs.UserSchema, error) {
	defaultUserMap, err := util.NewExpandingMap(constants.DEFAULT_USER_SCHEMA)
	if err != nil {
		return nil, fmt.Errorf("NewExpandingMap returns err: %v", err)
	}

	defaultUserSchema := &constructs.UserSchema{
		Schema: defaultUserMap,
	}

	return defaultUserSchema, dumpUserSchema(lsd.path, defaultUserSchema)
}

func dumpUserSchema(path string, schema *constructs.UserSchema) error {
	jsonBytes, err := json.Marshal(schema.Schema.ToRegularMap())
	if err != nil {
		return fmt.Errorf("json.Marshall(%+v) returns err: %w", schema, err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("os.WriteFile() returns err: %w", err)
	}

	return nil
}

func loadUserSchema(path string) (*constructs.UserSchema, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%s) returns err: %w", path, err)
	}

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll() returns err: %w", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(bytes, &schema); err != nil {
		return nil, fmt.Errorf("json.Unmarshal returns err: %w", err)
	}

	expandingSchema, err := util.NewExpandingMap(schema)
	if err != nil {
		return nil, fmt.Errorf("util.NewExpandingMap(%+v) returns err: %w", schema, err)
	}

	return &constructs.UserSchema{
		Schema: expandingSchema,
	}, nil
}
