package main

import (
	"testing"

	"github.com/davidramiro/fritzgandi/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	config.InitConfig()
	assert.NotNil(t, config.AppConfig, "init should fill config object")
}
