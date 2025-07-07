package ops

import (
	"errors"
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnecterrs "github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/stretchr/testify/require"
)

func TestErrorIsForbiddenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is ForbiddenError",
			err: &sdkkonnecterrs.ForbiddenError{
				Status:   403,
				Title:    "Quota Exceeded",
				Instance: "kong:trace:0000000000000000000",
				Detail:   "Maximum number of Active Networks exceeded. Max allowed: 0",
			},
			want: true,
		},
		{
			name: "error is SDKError with 403 status code",
			err: &sdkkonnecterrs.SDKError{
				StatusCode: 403,
				Body: `{
					"code": 7,
					"message": "usage constraint error",
					"details": [
						{
							"@type": "type.googleapis.com/kong.admin.model.v1.ErrorDetail",
							"messages": [
								"operation not permitted on KIC cluster"
							]
						}
					]
				}`,
			},
			want: true,
		},
		{
			name: "error is SDKError with non-403 status code",
			err: &sdkkonnecterrs.SDKError{
				StatusCode: 404,
				Body: `{
					"code": 7,
					"message": "usage constraint error",
					"details": [
						{
							"@type": "type.googleapis.com/kong.admin.model.v1.ErrorDetail",
							"messages": [
								"operation not permitted on KIC cluster"
							]
						}
					]
				}`,
			},
			want: false,
		},
		{
			name: "error is not SDKError",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorIsForbiddenError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestErrorIsSDKBadRequestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is BadRequestError",
			err:  &sdkkonnecterrs.BadRequestError{},
			want: true,
		},
		{
			name: "error is not BadRequestError",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorIsSDKBadRequestError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestErrorIsCreateConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is ConflictError",
			err:  &sdkkonnecterrs.ConflictError{},
			want: true,
		},
		{
			name: "error is SDKError with conflict message",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 3,
					"message": "data constraint error",
					"details": []
				}`,
			},
			want: true,
		},
		{
			name: "error is SDKError with non-conflict message",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 3,
					"message": "some other error",
					"details": []
				}`,
			},
			want: false,
		},
		{
			name: "error is SDKError with code 6",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 6,
					"message": "already exists",
					"details": []
				}`,
			},
			want: true,
		},
		{
			name: "error is SDKError with code 7",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 7,
					"message": "other error",
					"details": []
				}`,
			},
			want: false,
		},
		{
			name: "error is not ConflictError or SDKError",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorIsCreateConflict(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSDKErrorIsConflict(t *testing.T) {
	tests := []struct {
		name string
		err  *sdkkonnecterrs.SDKError
		want bool
	}{
		{
			name: "SDKError with data constraint error message and code 3",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 3,
					"message": "data constraint error",
					"details": []
				}`,
			},
			want: true,
		},
		{
			name: "SDKError with data constraint error message and code 6",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 6,
					"message": "data constraint error",
					"details": []
				}`,
			},
			want: true,
		},
		{
			name: "SDKError with unique constraint failed message",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 3,
					"message": "name (type: unique) constraint failed",
					"details": []
				}`,
			},
			want: true,
		},
		{
			name: "SDKError with non-conflict message",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 3,
					"message": "some other error",
					"details": []
				}`,
			},
			want: false,
		},
		{
			name: "SDKError with conflict message but different code",
			err: &sdkkonnecterrs.SDKError{
				Body: `{
					"code": 4,
					"message": "data constraint error",
					"details": []
				}`,
			},
			want: false,
		},
		{
			name: "SDKError with invalid JSON body",
			err: &sdkkonnecterrs.SDKError{
				Body: `invalid json`,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SDKErrorIsConflict(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestErrorIsDataPlaneGroupConflictProposedConfigIsTheSame(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not a ConflictError",
			err:  errors.New("not a conflict error"),
			want: false,
		},
		{
			name: "ConflictError with non-int code type",
			err: &sdkkonnecterrs.ConflictError{
				Status: 409.0,
				Detail: "Proposed configuration and current configuration are identical",
			},
			want: true,
		},
		{
			name: "ConflictError with wrong code",
			err: &sdkkonnecterrs.ConflictError{
				Status: 400,
				Detail: "Proposed configuration and current configuration are identical",
			},
			want: false,
		},
		{
			name: "ConflictError with non-string message",
			err: &sdkkonnecterrs.ConflictError{
				Status: 409,
				Detail: 12345,
			},
			want: false,
		},
		{
			name: "ConflictError with message missing expected substring",
			err: &sdkkonnecterrs.ConflictError{
				Status: 409,
				Detail: "Some other conflict detail",
			},
			want: false,
		},
		{
			name: "Valid ConflictError with matching float code and message",
			err: &sdkkonnecterrs.ConflictError{
				Status: 409.0,
				Detail: "Error: Proposed configuration and current configuration are identical, no changes required",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorIsDataPlaneGroupConflictProposedConfigIsTheSame(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestErrorIsDataPlaneGroupBadRequestPreviousConfigNotFinishedProvisioning(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is not BadRequestError",
			err:  errors.New("not a bad request error"),
			want: false,
		},
		{
			name: "BadRequestError with empty invalid parameters",
			err:  &sdkkonnecterrs.BadRequestError{InvalidParameters: []sdkkonnectcomp.InvalidParameters{}},
			want: false,
		},
		{
			name: "BadRequestError with invalid parameter of wrong type",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: "some_wrong_type",
					},
				},
			},
			want: false,
		},
		{
			name: "BadRequestError with matching type but nil InvalidParameterStandard",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
					},
				},
			},
			want: false,
		},
		{
			name: "BadRequestError with matching type, but non-matching field",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
						InvalidParameterStandard: &sdkkonnectcomp.InvalidParameterStandard{
							Field:  "wrong_field",
							Reason: "Data-plane groups in the previous configuration have not finished provisioning",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "BadRequestError with matching type and field but non-matching reason",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
						InvalidParameterStandard: &sdkkonnectcomp.InvalidParameterStandard{
							Field:  "dataplane_groups",
							Reason: "some other reason",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "BadRequestError with valid matching invalid parameter",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
						InvalidParameterStandard: &sdkkonnectcomp.InvalidParameterStandard{
							Field:  "dataplane_groups",
							Reason: "Error: Data-plane groups in the previous configuration have not finished provisioning",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "BadRequestError with multiple parameters where one matches",
			err: &sdkkonnecterrs.BadRequestError{
				InvalidParameters: []sdkkonnectcomp.InvalidParameters{
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
						InvalidParameterStandard: &sdkkonnectcomp.InvalidParameterStandard{
							Field:  "some_field",
							Reason: "irrelevant reason",
						},
					},
					{
						Type: sdkkonnectcomp.InvalidParametersTypeInvalidParameterStandard,
						InvalidParameterStandard: &sdkkonnectcomp.InvalidParameterStandard{
							Field:  "dataplane_groups",
							Reason: "Data-plane groups in the previous configuration have not finished provisioning fully",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorIsDataPlaneGroupBadRequestPreviousConfigNotFinishedProvisioning(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestErrorIsConflictError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "error is ConflictError with status 409",
			err: &sdkkonnecterrs.ConflictError{
				Status: 409.0,
				Detail: "Key (org_id, name) already exists.",
			},
			want: true,
		},
		{
			name: "error is ConflictError with non-409 status",
			err: &sdkkonnecterrs.ConflictError{
				Status: 400,
				Detail: "Some other error",
			},
			want: false,
		},
		{
			name: "error is not ConflictError",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "error is SDKError",
			err: &sdkkonnecterrs.SDKError{
				StatusCode: 409,
				Body:       "conflict error body",
			},
			want: false,
		},
		{
			name: "error is nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorIsConflictError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}
