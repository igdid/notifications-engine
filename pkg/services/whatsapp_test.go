package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend_WhatsApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.JSONEq(t, `{
			"to": "1234567890",
			"messaging_product": "whatsapp",
			"type": "document",
			"document": {
				"link": "https://example.com/file.pdf",
				"filename": "file",
				"caption": "message"
			}
		}`, string(b))
	}))
	defer ts.Close()

	service := NewWhatsAppService(WhatsAppOptions{
		ApiURL:             ts.URL,
		Token:              "token",
		InsecureSkipVerify: true,
	})
	err := service.Send(Notification{
		Message: "message",
		WhatsApp: &WhatsAppNotification{
			Type: "document",
			Document: &DocumentMessage{
				Link:     "https://example.com/file.pdf",
				Filename: "file",
				Caption:  "file",
			},
		},
	}, Destination{
		Service:   "whatsapp",
		Recipient: "1234567890",
	})
	require.NoError(t, err)
}
