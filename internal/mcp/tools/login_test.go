package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetConnectionInfo(t *testing.T) {
	tharsisURL := "https://tharsis.example.com"

	type testCase struct {
		name        string
		mockSetup   func(*mocks.CallerClient)
		expectAuth  bool
		expectTRN   *string
		profileName string
	}

	testCases := []testCase{
		{
			name:        "authenticated user",
			profileName: "default",
			mockSetup: func(cc *mocks.CallerClient) {
				cc.On("GetCaller", mock.Anything, &emptypb.Empty{}).Return(&pb.GetCallerResponse{
					Caller: &pb.GetCallerResponse_User{
						User: &pb.User{
							Metadata: &pb.ResourceMetadata{Trn: "trn:user:user1"},
							Username: "user1",
						},
					},
				}, nil)
			},
			expectAuth: true,
			expectTRN:  ptr.String("trn:user:user1"),
		},
		{
			name:        "authenticated service account",
			profileName: "default",
			mockSetup: func(cc *mocks.CallerClient) {
				cc.On("GetCaller", mock.Anything, &emptypb.Empty{}).Return(&pb.GetCallerResponse{
					Caller: &pb.GetCallerResponse_ServiceAccount{
						ServiceAccount: &pb.ServiceAccount{
							Metadata: &pb.ResourceMetadata{Trn: "trn:service_account:sa1"},
							Name:     "sa1",
						},
					},
				}, nil)
			},
			expectAuth: true,
			expectTRN:  ptr.String("trn:service_account:sa1"),
		},
		{
			name:        "not authenticated",
			profileName: "default",
			mockSetup: func(cc *mocks.CallerClient) {
				cc.On("GetCaller", mock.Anything, &emptypb.Empty{}).Return(nil, status.Error(codes.Unauthenticated, "not authenticated"))
			},
			expectAuth: false,
			expectTRN:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCaller := mocks.NewCallerClient(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockCaller)
			}

			toolCtx := &ToolContext{
				tharsisURL:  tharsisURL,
				profileName: tc.profileName,
				grpcClient:  &client.Client{CallerClient: mockCaller},
			}

			_, handler := getConnectionInfo(toolCtx)
			_, output, err := handler(t.Context(), nil, &getConnectionInfoInput{})

			require.NoError(t, err)
			assert.Equal(t, tharsisURL, output.TharsisURL)
			assert.Equal(t, tc.profileName, output.ProfileName)
			assert.Equal(t, tc.expectAuth, output.Authenticated)
			if tc.expectTRN != nil {
				require.NotNil(t, output.TRN)
				assert.Equal(t, *tc.expectTRN, *output.TRN)
			} else {
				assert.Nil(t, output.TRN)
			}
		})
	}
}

func TestLoginWithSSO(t *testing.T) {
	type testCase struct {
		name        string
		mockSetup   func(*auth.MockAuthenticator)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "successful login",
			mockSetup: func(ma *auth.MockAuthenticator) {
				token := &oauth2.Token{AccessToken: "test-token"}
				ma.On("PerformLogin", mock.Anything).Return(token, nil)
				ma.On("StoreToken", token).Return(nil)
			},
		},
		{
			name: "login fails",
			mockSetup: func(ma *auth.MockAuthenticator) {
				ma.On("PerformLogin", mock.Anything).Return(nil, status.Error(codes.Unauthenticated, "login failed"))
			},
			expectError: true,
		},
		{
			name: "store token fails",
			mockSetup: func(ma *auth.MockAuthenticator) {
				token := &oauth2.Token{AccessToken: "test-token"}
				ma.On("PerformLogin", mock.Anything).Return(token, nil)
				ma.On("StoreToken", token).Return(status.Error(codes.Internal, "store failed"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuth := auth.NewMockAuthenticator(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockAuth)
			}

			toolCtx := &ToolContext{
				authenticator: mockAuth,
			}

			_, handler := loginWithSSO(toolCtx)
			_, output, err := handler(t.Context(), nil, &ssoLoginInput{})

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
			assert.Contains(t, output.Message, "Successfully logged in")
		})
	}
}
