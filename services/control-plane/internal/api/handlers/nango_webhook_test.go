package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVerifyNangoWebhook(t *testing.T) {
	body := []byte(`{"type":"auth","success":true,"connectionId":"abc","providerConfigKey":"pck","tags":{"end_user_id":"tenant-1"}}`)
	webhookSecret := "test-webhook-signing-key"
	apiSecret := "test-api-secret-key"

	hmacFor := func(secret string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		return hex.EncodeToString(mac.Sum(nil))
	}
	legacyFor := func(secret string) string {
		combined := append([]byte(secret), body...)
		sum := sha256.Sum256(combined)
		return hex.EncodeToString(sum[:])
	}

	tests := []struct {
		name          string
		webhookSecret string
		apiSecret     string
		headers       map[string]string
		want          bool
	}{
		{
			name:          "nango hmac header with api secret",
			webhookSecret: webhookSecret,
			apiSecret:     apiSecret,
			headers:       map[string]string{"X-Nango-Hmac-Sha256": hmacFor(apiSecret)},
			want:          true,
		},
		{
			name:          "legacy signature header with api secret",
			webhookSecret: webhookSecret,
			apiSecret:     apiSecret,
			headers:       map[string]string{"X-Nango-Signature": legacyFor(apiSecret)},
			want:          true,
		},
		{
			name:          "hmac on signature header with webhook signing key",
			webhookSecret: webhookSecret,
			apiSecret:     apiSecret,
			headers:       map[string]string{"X-Nango-Signature": hmacFor(webhookSecret)},
			want:          true,
		},
		{
			name:          "wrong secret",
			webhookSecret: webhookSecret,
			apiSecret:     apiSecret,
			headers:       map[string]string{"X-Nango-Hmac-Sha256": hmacFor("wrong")},
			want:          false,
		},
		{
			name:          "no headers",
			webhookSecret: webhookSecret,
			apiSecret:     apiSecret,
			headers:       map[string]string{},
			want:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hdr := make(http.Header)
			for k, v := range tc.headers {
				hdr.Set(k, v)
			}
			got := verifyNangoWebhook(body, tc.webhookSecret, tc.apiSecret, hdr)
			if got != tc.want {
				t.Fatalf("verifyNangoWebhook() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNangoWebhookPayloadTenantID(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "tags end_user_id",
			raw:  `{"tags":{"end_user_id":"tenant-from-tags"}}`,
			want: "tenant-from-tags",
		},
		{
			name: "endUser endUserId",
			raw:  `{"endUser":{"endUserId":"tenant-from-end-user"}}`,
			want: "tenant-from-end-user",
		},
		{
			name: "legacy endUser id",
			raw:  `{"endUser":{"id":"tenant-legacy"}}`,
			want: "tenant-legacy",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p nangoWebhookPayload
			if err := json.Unmarshal([]byte(tc.raw), &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := p.tenantID(); got != tc.want {
				t.Fatalf("tenantID() = %q, want %q", got, tc.want)
			}
		})
	}
}
