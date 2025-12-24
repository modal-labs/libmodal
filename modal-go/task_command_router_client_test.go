package modal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mockLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func mockJWT(exp any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	var payloadJSON []byte
	if exp != nil {
		payloadJSON, _ = json.Marshal(map[string]any{"exp": exp})
	} else {
		payloadJSON, _ = json.Marshal(map[string]any{})
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := "fake-signature"
	return header + "." + payload + "." + signature
}

func TestParseJwtExpirationWithValidJWT(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	exp := time.Now().Unix() + 3600
	jwt := mockJWT(exp)
	result := parseJwtExpiration(context.Background(), jwt, mockLogger())
	g.Expect(result).ToNot(gomega.BeNil())
	g.Expect(*result).To(gomega.Equal(exp))
}

func TestParseJwtExpirationWithoutExpClaim(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	jwt := mockJWT(nil)
	result := parseJwtExpiration(context.Background(), jwt, mockLogger())
	g.Expect(result).To(gomega.BeNil())
}

func TestParseJwtExpirationWithMalformedJWT(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	jwt := "only.two"
	result := parseJwtExpiration(context.Background(), jwt, mockLogger())
	g.Expect(result).To(gomega.BeNil())
}

func TestParseJwtExpirationWithInvalidBase64(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	jwt := "invalid.!!!invalid!!!.signature"
	result := parseJwtExpiration(context.Background(), jwt, mockLogger())
	g.Expect(result).To(gomega.BeNil())
}

func TestParseJwtExpirationWithNonNumericExp(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	jwt := mockJWT("not-a-number")
	result := parseJwtExpiration(context.Background(), jwt, mockLogger())
	g.Expect(result).To(gomega.BeNil())
}

func TestCallWithRetriesOnTransientErrorsSuccessOnFirstAttempt(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	callCount := 0
	result, err := callWithRetriesOnTransientErrors(ctx, func() (*string, error) {
		callCount++
		output := "success"
		return &output, nil
	}, defaultRetryOptions())

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(*result).To(gomega.Equal("success"))
	g.Expect(callCount).To(gomega.Equal(1))
}

func TestCallWithRetriesOnTransientErrorsRetriesOnTransientCodes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		code    codes.Code
		message string
	}{
		{"DeadlineExceeded", codes.DeadlineExceeded, "timeout"},
		{"Unavailable", codes.Unavailable, "unavailable"},
		{"Canceled", codes.Canceled, "cancelled"},
		{"Internal", codes.Internal, "internal error"},
		{"Unknown", codes.Unknown, "unknown error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)
			ctx := context.Background()
			callCount := 0
			result, err := callWithRetriesOnTransientErrors(ctx, func() (*string, error) {
				callCount++
				var output string
				if callCount == 1 {
					output = ""
					return &output, status.Error(tc.code, tc.message)
				}
				output = "success"
				return &output, nil
			}, retryOptions{BaseDelay: time.Millisecond, DelayFactor: 1, MaxRetries: intPtr(10)})

			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(*result).To(gomega.Equal("success"))
			g.Expect(callCount).To(gomega.Equal(2))
		})
	}
}

func TestCallWithRetriesOnTransientErrorsNonRetryableError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	callCount := 0
	_, err := callWithRetriesOnTransientErrors(ctx, func() (*string, error) {
		callCount++
		return nil, status.Error(codes.InvalidArgument, "invalid")
	}, retryOptions{BaseDelay: time.Millisecond, DelayFactor: 1, MaxRetries: intPtr(10)})

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(callCount).To(gomega.Equal(1))
}

func TestCallWithRetriesOnTransientErrorsMaxRetriesExceeded(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	callCount := 0
	maxRetries := 3
	_, err := callWithRetriesOnTransientErrors(ctx, func() (*string, error) {
		callCount++
		return nil, status.Error(codes.Unavailable, "unavailable")
	}, retryOptions{BaseDelay: time.Millisecond, DelayFactor: 1, MaxRetries: &maxRetries})

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(callCount).To(gomega.Equal(maxRetries + 1))
}

func TestCallWithRetriesOnTransientErrorsDeadlineExceeded(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()
	callCount := 0
	deadline := time.Now().Add(50 * time.Millisecond)
	_, err := callWithRetriesOnTransientErrors(ctx, func() (*string, error) {
		callCount++
		return nil, status.Error(codes.Unavailable, "unavailable")
	}, retryOptions{BaseDelay: 100 * time.Millisecond, DelayFactor: 1, MaxRetries: nil, Deadline: &deadline})

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("deadline exceeded"))
}

func intPtr(i int) *int {
	return &i
}
