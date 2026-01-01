package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dbAndUserTestCase struct {
	dbnameOpt string
	userOpt   string
	argDB     string
	argUser   string

	expectedDB   string
	expectedUser string
}

func TestResolveDBAndUser(t *testing.T) {
	testcases := []struct {
		name	 string
		input	 dbAndUserTestCase
	} {
		{
			name: "Both flags provided, args ignored",
			input: dbAndUserTestCase{
				dbnameOpt: "flagDB",
				userOpt:   "flagUser",
				argDB:     "argDB",
				argUser:   "argUser",
				expectedDB:   "flagDB",
				expectedUser: "flagUser",
			},
		},
		{
			name: "Only dbname flag provided",
			input: dbAndUserTestCase{
				dbnameOpt: "flagDB",
				userOpt:   "",
				argDB:     "argDB",
				argUser:   "argUser",
				expectedDB:   "flagDB",
				expectedUser: "argUser",
			},
		},
		{
			name: "Only username flag provided",
			input: dbAndUserTestCase{
				dbnameOpt: "",
				userOpt:   "flagUser",
				argDB:     "argDB",
				argUser:   "argUser",
				expectedDB:   "argDB",
				expectedUser: "flagUser",
			},
		},
		{
			name: "No flags provided, args used",
			input: dbAndUserTestCase{
				dbnameOpt: "",
				userOpt:   "",
				argDB:     "argDB",
				argUser:   "argUser",
				expectedDB:   "argDB",
				expectedUser: "argUser",
			},
		},
		{
			name: "No flags or args provided",
			input: dbAndUserTestCase{
				dbnameOpt: "",
				userOpt:   "",
				argDB:     "",
				argUser:   "",
				expectedDB:   "",
				expectedUser: "",
			},
		},
		{
			name: "Only dbname flag and argDB provided",
			input: dbAndUserTestCase{
				dbnameOpt: "flagDB",
				userOpt:   "",
				argDB:     "argDB",
				argUser:   "",
				expectedDB:   "flagDB",
				expectedUser: "argDB",
			},
		},
		{
			name: "Only username flag and argUser provided",
			input: dbAndUserTestCase{
				dbnameOpt: "",
				userOpt:   "flagUser",
				argDB:     "",
				argUser:   "argUser",
				expectedDB:   "",
				expectedUser: "flagUser",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualDB, actualUser := resolveDBAndUser(tc.input.dbnameOpt, tc.input.userOpt, tc.input.argDB, tc.input.argUser)
			assert.Equal(t, tc.input.expectedDB, actualDB, "finalDB does not match expected value")
			assert.Equal(t, tc.input.expectedUser, actualUser, "finalUser does not match expected value")
		})
	}	
}
