package controllers

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"main/Models"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type MockDB struct {
	mock.Mock
}

// Implement Query method for MockDB
func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	argsList := m.Called(append([]interface{}{query}, args...)...)
	if result, ok := argsList.Get(0).(*sql.Rows); ok {
		return result, argsList.Error(1)
	}
	return nil, argsList.Error(1)
}

func TestGetUserRoles(t *testing.T) {
	type testCase struct {
		name        string
		email       string
		mockData    [][]interface{}
		expectedLen int
		mockError   error
	}

	testCases := []testCase{
		{
			name:  "success - email filter",
			email: "test@example.com",
			mockData: [][]interface{}{
				{1, "test@example.com", 2, time.Now(), time.Now(), nil, "admin"},
			},
			expectedLen: 1,
		},
		{
			name:  "success - all users",
			email: "",
			mockData: [][]interface{}{
				{1, "test@example.com", 2, time.Now(), time.Now(), nil, "admin"},
				{2, "user@example.com", 3, time.Now(), time.Now(), nil, "user"},
			},
			expectedLen: 2,
		},
		{
			name:        "no users found",
			email:       "notfound@example.com",
			mockData:    [][]interface{}{}, // No data
			expectedLen: 0,
		},
		{
			name:        "database error",
			email:       "",
			mockData:    nil,
			expectedLen: 0,
			mockError:   errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			// Define expected query behavior
			query := `SELECT user_roles.id, user_roles.email, user_roles.role_id, user_roles.created_at, user_roles.updated_at,
				user_roles.deleted_at, roles.role_key 
				FROM user_roles 
				LEFT JOIN roles ON user_roles.role_id = roles.id  
				WHERE user_roles.deleted_at IS NULL`

			if tc.email != "" {
				query += ` AND user_roles.email = ?`
			}

			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "email", "role_id", "created_at", "updated_at", "deleted_at", "role_key"})
				for _, row := range tc.mockData {
					// Convert row to driver.Values and add to rows
					var values []driver.Value
					for _, v := range row {
						values = append(values, v)
					}
					rows.AddRow(values...)
				}
				mock.ExpectQuery(query).WillReturnRows(rows)
			}

			// Create test HTTP request
			url := "/roles"
			if tc.email != "" {
				url += "?email=" + tc.email
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler := GetUserRoles(db)
			handler.ServeHTTP(w, req)

			// Assert HTTP response
			if tc.mockError != nil {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				// Decode response
				var roles []models.UserRole
				json.NewDecoder(w.Body).Decode(&roles)
				assert.Len(t, roles, tc.expectedLen)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Test GetUserRole

func TestGetUserRole(t *testing.T) {
	type testCase struct {
		name      string
		userID    string
		mockData  []interface{}
		expectErr bool
		mockError error
	}

	testCases := []testCase{
		{
			name:   "success - valid user",
			userID: "1",
			mockData: []interface{}{
				1, "test@example.com", 2, time.Now(), time.Now(), nil, "admin",
			},
			expectErr: false,
		},
		{
			name:      "user not found",
			userID:    "99",
			mockData:  nil,
			expectErr: true,
		},
		{
			name:      "database error",
			userID:    "1",
			mockData:  nil,
			mockError: errors.New("database error"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			query := regexp.QuoteMeta(`SELECT user_roles.id, user_roles.email, user_roles.role_id, 
				user_roles.created_at, user_roles.updated_at, user_roles.deleted_at, roles.role_key 
				FROM user_roles 
				LEFT JOIN roles ON user_roles.role_id = roles.id  
				WHERE user_roles.id = $1 AND user_roles.deleted_at IS NULL`)

			if tc.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tc.userID).WillReturnError(tc.mockError)
			} else if tc.mockData != nil {
				rowValues := make([]driver.Value, len(tc.mockData))
				for i, v := range tc.mockData {
					rowValues[i] = v
				}

				rows := sqlmock.NewRows([]string{"id", "email", "role_id", "created_at", "updated_at", "deleted_at", "role_key"}).
					AddRow(rowValues...)

				mock.ExpectQuery(query).WithArgs(tc.userID).WillReturnRows(rows).RowsWillBeClosed()
			}

			req := httptest.NewRequest("GET", "/roles/"+tc.userID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.userID})

			handler := GetUserRole(db)
			handler.ServeHTTP(w, req)

			// Debug response body
			fmt.Println("Response Code:", w.Code)
			fmt.Println("Response Body:", w.Body.String())

			if tc.expectErr {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				var userRole models.UserRole
				err := json.NewDecoder(w.Body).Decode(&userRole)
				assert.NoError(t, err)
				assert.Equal(t, tc.mockData[1], userRole.Email)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Test CreateUserRole

func TestCreateUserRole(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    testCases := []struct {
        name         string
        requestBody  string
        expectedCode int
        mockQueries  func()
    }{
        {
            name:         "success - valid request",
            requestBody:  `{"email": "test@example.com", "role_id": 2}`,
            expectedCode: http.StatusCreated,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(2).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

                mock.ExpectQuery(`INSERT INTO user_roles \(email, role_id\) VALUES \(\$1, \$2\) RETURNING id, created_at, updated_at`).
                    WithArgs("test@example.com", 2).
                    WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
                        AddRow(1, time.Now(), time.Now()))
            },
        },
        {
            name:         "failure - role does not exist",
            requestBody:  `{"email": "test@example.com", "role_id": 2}`,
            expectedCode: http.StatusBadRequest,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(2).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
            },
        },
        {
            name:         "failure - invalid JSON",
            requestBody:  `{"email": "test@example.com", "role_id":}`,
            expectedCode: http.StatusBadRequest,
            mockQueries: func() {
                // No DB queries should run because JSON is invalid
            },
        },
        {
            name:         "failure - database error on insert",
            requestBody:  `{"email": "test@example.com", "role_id": 2}`,
            expectedCode: http.StatusInternalServerError,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(2).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

                mock.ExpectQuery(`INSERT INTO user_roles \(email, role_id\) VALUES \(\$1, \$2\) RETURNING id, created_at, updated_at`).
                    WithArgs("test@example.com", 2).
                    WillReturnError(errors.New("insert error"))
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            tc.mockQueries()

            req := httptest.NewRequest("POST", "/user_roles", strings.NewReader(tc.requestBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            handler := CreateUserRole(db)
            handler.ServeHTTP(w, req)

            assert.Equal(t, tc.expectedCode, w.Code)

            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}


func TestUpdateUserRole(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    testCases := []struct {
        name         string
        userID       string
        requestBody  string
        expectedCode int
        mockQueries  func()
    }{
        {
            name:         "success - valid request",
            userID:       "1",
            requestBody:  `{"email": "updated@example.com", "role_id": 2}`,
            expectedCode: http.StatusOK,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(2).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

                mock.ExpectExec(`UPDATE user_roles SET email = \$1, role_id = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$3 AND deleted_at IS NULL`).
                    WithArgs("updated@example.com", 2, "1").
                    WillReturnResult(sqlmock.NewResult(1, 1))
            },
        },
        {
            name:         "failure - role does not exist",
            userID:       "1",
            requestBody:  `{"email": "updated@example.com", "role_id": 99}`,
            expectedCode: http.StatusBadRequest,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(99).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
            },
        },
        {
            name:         "failure - invalid JSON",
            userID:       "1",
            requestBody:  `{"email": "updated@example.com", "role_id":}`,
            expectedCode: http.StatusBadRequest,
            mockQueries: func() {},
        },
        {
            name:         "failure - database error on update",
            userID:       "1",
            requestBody:  `{"email": "updated@example.com", "role_id": 2}`,
            expectedCode: http.StatusInternalServerError,
            mockQueries: func() {
                mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM roles WHERE id = \$1 AND deleted_at IS NULL\)`).
                    WithArgs(2).
                    WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

                mock.ExpectExec(`UPDATE user_roles SET email = \$1, role_id = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$3 AND deleted_at IS NULL`).
                    WithArgs("updated@example.com", 2, "1").
                    WillReturnError(errors.New("update error"))
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            tc.mockQueries()

            req := httptest.NewRequest("PUT", "/user_roles/"+tc.userID, strings.NewReader(tc.requestBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            req = mux.SetURLVars(req, map[string]string{"id": tc.userID})

            handler := UpdateUserRole(db)
            handler.ServeHTTP(w, req)

            assert.Equal(t, tc.expectedCode, w.Code)
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}


func TestDeleteUserRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		userID       int // Now an int
		expectedCode int
		mockExec     func()
	}{
		{
			name:         "success - user role deleted",
			userID:       1,
			expectedCode: http.StatusNoContent,
			mockExec: func() {
				mock.ExpectExec(`UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
			},
		},
		{
			name:         "failure - user role not found",
			userID:       99,
			expectedCode: http.StatusNotFound,
			mockExec: func() {
				mock.ExpectExec(`UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
		},
		{
			name:         "failure - database error",
			userID:       1,
			expectedCode: http.StatusInternalServerError,
			mockExec: func() {
				mock.ExpectExec(`UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockExec()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/user_roles/%d", tc.userID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": fmt.Sprintf("%d", tc.userID)})
			w := httptest.NewRecorder()

			handler := DeleteUserRole(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			err := mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}





