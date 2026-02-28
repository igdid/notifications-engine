package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	texttemplate "text/template"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type MessageType string

const (
	apiVersion string = "v25.0"
)

const (
	TypeText        MessageType = "text"
	TypeImage       MessageType = "image"
	TypeDocument    MessageType = "document"
	TypeInteractive MessageType = "interactive"
	TypeTemplate    MessageType = "template"
)

type WhatsAppNotification struct {
	MessagingProduct string           `json:"messaging_product"`
	To               string           `json:"to"`
	Type             MessageType      `json:"type"`
	Text             *TextMessage     `json:"text,omitempty"`
	Image            *ImageMessage    `json:"image,omitempty"`
	Document         *DocumentMessage `json:"document,omitempty"`
	Interactive      *Interactive     `json:"interactive,omitempty"`
	Template         *TemplateMessage `json:"template,omitempty"`
}

func (n *WhatsAppNotification) GetTemplater(_ string, _ texttemplate.FuncMap) (Templater, error) {
	return func(_ *Notification, _ map[string]any) error {
		return nil
	}, nil
}

type TextMessage struct {
	PreviewURL bool   `json:"preview_url,omitempty"`
	Body       string `json:"body"`
}

type ImageMessage struct {
	Link    string `json:"link,omitempty"`
	ID      string `json:"id,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type DocumentMessage struct {
	Link     string `json:"link,omitempty"`
	ID       string `json:"id,omitempty"`
	Filename string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type Interactive struct {
	Type   string            `json:"type"` // "button"
	Body   InteractiveBody   `json:"body"`
	Action InteractiveAction `json:"action"`
}

type InteractiveBody struct {
	Text string `json:"text"`
}

type InteractiveAction struct {
	Buttons []InteractiveButton `json:"buttons"`
}

type InteractiveButton struct {
	Type  string      `json:"type"` // "reply"
	Reply ButtonReply `json:"reply"`
}

type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type TemplateMessage struct {
	Name       string              `json:"name"`
	Language   TemplateLanguage    `json:"language"`
	Components []TemplateComponent `json:"components,omitempty"`
}

type TemplateLanguage struct {
	Code string `json:"code"` // "ru", "en_US", etc.
}

type TemplateComponent struct {
	Type       string              `json:"type"`               // "body", "header", "button"
	SubType    string              `json:"sub_type,omitempty"` // для кнопок
	Index      string              `json:"index,omitempty"`    // для кнопок
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

type TemplateParameter struct {
	Type     string            `json:"type"` // "text", "image", "document", "currency", "date_time"
	Text     string            `json:"text,omitempty"`
	Image    *TemplateMedia    `json:"image,omitempty"`
	Document *TemplateMedia    `json:"document,omitempty"`
	Currency *TemplateCurrency `json:"currency,omitempty"`
	DateTime *TemplateDateTime `json:"date_time,omitempty"`
}

type TemplateMedia struct {
	Link string `json:"link,omitempty"`
	ID   string `json:"id,omitempty"`
}

type TemplateCurrency struct {
	FallbackValue string `json:"fallback_value"`
	Code          string `json:"code"`
	Amount1000    int    `json:"amount_1000"`
}

type TemplateDateTime struct {
	FallbackValue string `json:"fallback_value"`
}

func NewTextMessage(to, body string, preview bool) *WhatsAppNotification {
	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeText,
		Text: &TextMessage{
			PreviewURL: preview,
			Body:       body,
		},
	}
}

func NewImageMessageByURL(to, url, caption string) *WhatsAppNotification {
	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeImage,
		Image: &ImageMessage{
			Link:    url,
			Caption: caption,
		},
	}
}

func NewDocumentMessageByURL(to, url, filename, caption string) *WhatsAppNotification {
	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeDocument,
		Document: &DocumentMessage{
			Link:     url,
			Filename: filename,
			Caption:  caption,
		},
	}
}

func NewButtonMessage(to, body string, buttons []ButtonReply) *WhatsAppNotification {
	btns := make([]InteractiveButton, len(buttons))
	for i, b := range buttons {
		btns[i] = InteractiveButton{
			Type:  "reply",
			Reply: b,
		}
	}

	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeInteractive,
		Interactive: &Interactive{
			Type: "button",
			Body: InteractiveBody{
				Text: body,
			},
			Action: InteractiveAction{
				Buttons: btns,
			},
		},
	}
}

func NewTemplateWithBodyParams(
	to, name, lang string,
	params ...string,
) *WhatsAppNotification {
	parameters := make([]TemplateParameter, len(params))
	for i, p := range params {
		parameters[i] = TemplateParameter{
			Type: "text",
			Text: p,
		}
	}

	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeTemplate,
		Template: &TemplateMessage{
			Name: name,
			Language: TemplateLanguage{
				Code: lang,
			},
			Components: []TemplateComponent{
				{
					Type:       "body",
					Parameters: parameters,
				},
			},
		},
	}
}

type WhatsAppOptions struct {
	BusinessPhoneNumber string `json:"businessPhoneNumber"`
	Token               string `json:"token"`
	ApiURL              string `json:"apiUrl"`
	InsecureSkipVerify  bool   `json:"insecureSkipVerify"`
	httputil.TransportOptions
}

type whatsappService struct {
	opts WhatsAppOptions
}

func NewWhatsAppService(opts WhatsAppOptions) NotificationService {
	return &whatsappService{opts: opts}
}

func (m *whatsappService) Send(notification Notification, dest Destination) (err error) {
	client, err := httputil.NewServiceHTTPClient(m.opts.TransportOptions, m.opts.InsecureSkipVerify, m.opts.ApiURL, "whatsapp")
	if err != nil {
		return err
	}
	var wn *WhatsAppNotification
	if notification.WhatsApp == nil {
		wn = NewTextMessage(dest.Recipient, notification.Message, true)
	} else {
		wn = notification.WhatsApp
		if notification.Message != "" {
			switch wn.Type {
			case "text":
				wn.Text.Body = notification.Message
			case "image":
				wn.Image.Caption = notification.Message
			case "document":
				wn.Document.Caption = notification.Message
			case "interactive":
				wn.Interactive.Body.Text = notification.Message
			case "template":
				wn.Template.Name = notification.Message
			default:
				return fmt.Errorf("unknown type: %s", wn.Type)
			}
		}
	}

	b, _ := json.Marshal(wn)

	uri := fmt.Sprintf("/%s/%s/messages", apiVersion, m.opts.BusinessPhoneNumber)
	req, err := http.NewRequest(http.MethodPost, m.opts.ApiURL+uri, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.opts.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("request to %s has failed with error code %d : %s", string(b), res.StatusCode, string(data))
	}

	return nil
}
