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

func TestGetTemplater_WhatsApp(t *testing.T) {
	n := Notification{
		Message: "message",
		WhatsApp: &WhatsAppNotification{
			Type: "{{.type}}",
			Document: &DocumentMessage{
				Link:     "https://example.com/{{.user_id}}.pdf",
				ID:       "{{.document_id}}",
				Filename: "file",
			},
			Image: &ImageMessage{
				Link: "https://example.com/{{.user_id}}.jpg",
				ID:   "{{.image_id}}",
			},
		},
	}
	templater, err := n.GetTemplater("", template.FuncMap{})

	require.NoError(t, err)

	var notification Notification
	err = templater(&notification, map[string]any{
		"type":     "image",
		"user_id":  "123456",
		"image_id": "987654",
	})

	require.NoError(t, err)

	assert.Equal(t, TypeImage, notification.WhatsApp.Type)
	assert.Equal(t, "https://example.com/123456.jpg", notification.WhatsApp.Image.Link)
	assert.Equal(t, "987654", notification.WhatsApp.Image.ID)
	err = templater(&notification, map[string]any{
		"type":        "document",
		"user_id":     "123456",
		"document_id": "987654",
	})

	require.NoError(t, err)

	assert.Equal(t, TypeDocument, notification.WhatsApp.Type)
	assert.Equal(t, "https://example.com/123456.pdf", notification.WhatsApp.Document.Link)
	assert.Equal(t, "987654", notification.WhatsApp.Document.ID)
}
