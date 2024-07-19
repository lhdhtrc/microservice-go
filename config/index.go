package config

import (
	loger "github.com/lhdhtrc/logger-go/pkg"
	tpg "github.com/lhdhtrc/task-go/pkg"
)

type CoreEntity struct {
	System SystemConfigEntity `json:"system" bson:"system" yaml:"system" mapstructure:"system"`
	Logger loger.ConfigEntity `json:"logger" bson:"logger" yaml:"logger" mapstructure:"logger"`
	Task   tpg.ConfigEntity   `json:"task" bson:"task" yaml:"task" mapstructure:"task"`
}
