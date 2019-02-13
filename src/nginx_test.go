package main

import (
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntityLocal(t *testing.T) {
	args = argumentList{
		RemoteMonitoring: false,
	}
	i, err := integration.New("test", integrationVersion)
	assert.NoError(t, err)

	e, err := entity(i)
	assert.NoError(t, err)
	assert.Nil(t, e.Metadata)
}

func TestParseURL(t *testing.T) {
	hostname1, port1, err1 := parseStatusURL("http://localhost/status")
	assert.NoError(t, err1)
	assert.Equal(t, "localhost", hostname1)
	assert.Equal(t, "80", port1)

	hostname2, port2, err2 := parseStatusURL("https://localhost/status")
	assert.NoError(t, err2)
	assert.Equal(t, "localhost", hostname2)
	assert.Equal(t, "443", port2)

	hostname3, port3, err3 := parseStatusURL("https://localhost:1234/status")
	assert.NoError(t, err3)
	assert.Equal(t, "localhost", hostname3)
	assert.Equal(t, "1234", port3)

	_, _, err4 := parseStatusURL("://localhost/status")
	assert.EqualError(t, err4, "parse ://localhost/status: missing protocol scheme")

	_, _, err5 := parseStatusURL("localhost/status")
	assert.EqualError(t, err5, "unsupported protocol scheme")
}

func TestEntityRemote(t *testing.T) {
	args = argumentList{
		StatusURL:        "http://test:1234/status",
		RemoteMonitoring: true,
	}
	i, err := integration.New("test", integrationVersion)
	assert.NoError(t, err)

	e, err := entity(i)
	assert.NoError(t, err)
	assert.Equal(t, "test:1234", e.Metadata.Name)
	assert.Equal(t, entityRemoteType, e.Metadata.Namespace)
}
