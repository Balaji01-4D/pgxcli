package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPGConnector_UpdatePassword_ConnString(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		connString string
		// Named input parameters for target function.
		newPassword string
	}{
		{
			name:        "uri format connection string",
			connString:  "postgresql://testuser:oldpassword@localhost:5432/testdb",
			newPassword: "newpassword123",
		},
		{
			name:        "uri format coonnection string when no password initially",
			connString:  "postgresql://testuser@localhost:5432/testdb",
			newPassword: "newpassword123",
		},
		{
			name:        "Update password in connection string",
			connString:  "host=localhost port=5432 user=testuser dbname=testdb password=oldpassword",
			newPassword: "newpassword123",
		},
		{
			name:        "Update password when no password initially",
			connString:  "host=localhost port=5432 user=testuser dbname=testdb",
			newPassword: "securepassword",
		},
		{
			name:        "Update empty password to another empty password",
			connString:  "host=localhost port=5432 user=testuser dbname=testdb password=",
			newPassword: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewPGConnectorFromConnString(tt.connString)
			assert.NoError(t, err)
			c.UpdatePassword(tt.newPassword)

			assert.Equal(t, tt.newPassword, c.cfg.Password)
		})
	}
}

func TestPGConnector_UpdatePassword_Fields(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		host        string
		database    string
		user        string
		password    string
		port        uint16
		newPassword string
	}{
		{
			name:        "Update password in fields",
			host:        "localhost",
			database:    "testdb",
			user:        "testuser",
			password:    "oldpassword",
			port:        5432,
			newPassword: "newpassword123",
		},
		{
			name:        "Update empty password to another password",
			host:        "localhost",
			database:    "testdb",
			user:        "testuser",
			password:    "",
			port:        5432,
			newPassword: "securepassword",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewPGConnectorFromFields(tt.host, tt.database, tt.user, tt.password, tt.port)
			assert.NoError(t, err)
			c.UpdatePassword(tt.newPassword)

			assert.Equal(t, tt.newPassword, c.cfg.Password)
		})
	}
}
