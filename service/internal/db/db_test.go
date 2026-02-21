package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	dbmdl "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/yamlfiles"
	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func TestGetNewDatabaseConnection(t *testing.T) {
	origDriver := RuntimeConfig.Database.Driver
	defer func() { RuntimeConfig.Database.Driver = origDriver }()

	RuntimeConfig.Database.Driver = "dummy"
	conn := GetNewDatabaseConnection()
	require.NotNil(t, conn)
	_, ok := conn.(*dummy.Dummy)
	assert.True(t, ok, "expected *dummy.Dummy for driver=dummy")

	RuntimeConfig.Database.Driver = "todotxt"
	conn = GetNewDatabaseConnection()
	require.NotNil(t, conn)
	_, ok = conn.(*todotxt.TodoTxt)
	assert.True(t, ok, "expected *todotxt.TodoTxt for driver=todotxt")

	RuntimeConfig.Database.Driver = "yamlfiles"
	conn = GetNewDatabaseConnection()
	require.NotNil(t, conn)
	_, ok = conn.(*yamlfiles.YamlFileDriver)
	assert.True(t, ok, "expected *yamlfiles.YamlFileDriver for driver=yamlfiles")

	RuntimeConfig.Database.Driver = "unknown"
	conn = GetNewDatabaseConnection()
	require.NotNil(t, conn)
	_, ok = conn.(*dummy.Dummy)
	assert.True(t, ok, "unknown driver should fall back to dummy")
}

func TestGetNewDatabaseConnection_ImplementsDB(t *testing.T) {
	origDriver := RuntimeConfig.Database.Driver
	defer func() { RuntimeConfig.Database.Driver = origDriver }()

	RuntimeConfig.Database.Driver = "dummy"
	conn := GetNewDatabaseConnection()
	require.Implements(t, (*dbmdl.DB)(nil), conn)
}
