package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	texttemplate "text/template"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type MessageType string

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

func (n *WhatsAppNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	msgType, err := texttemplate.New(name).Funcs(f).Parse(string(n.Type))
	if err != nil {
		return nil, err
	}
	return func(notification *Notification, vars map[string]any) error {
		if notification.WhatsApp == nil {
			notification.WhatsApp = &WhatsAppNotification{}
		}
		var msgTypeData bytes.Buffer
		if err := msgType.Execute(&msgTypeData, vars); err != nil {
			return err
		}
		notification.WhatsApp.Type = MessageType(msgTypeData.String())

		switch notification.WhatsApp.Type {
		case "image":
			link, err := texttemplate.New(name).Funcs(f).Parse(n.Image.Link)
			if err != nil {
				return err
			}
			var linkData bytes.Buffer
			if err := link.Execute(&linkData, vars); err != nil {
				return err
			}
			if notification.WhatsApp.Image == nil {
				notification.WhatsApp.Image = &ImageMessage{}
			}
			notification.WhatsApp.Image.Link = linkData.String()
			id, err := texttemplate.New(name).Funcs(f).Parse(n.Image.ID)
			if err != nil {
				return err
			}
			var idData bytes.Buffer
			if err := id.Execute(&idData, vars); err != nil {
				return err
			}
			notification.WhatsApp.Image.ID = idData.String()
		case "document":
			link, err := texttemplate.New(name).Funcs(f).Parse(n.Document.Link)
			if err != nil {
				return err
			}
			var linkData bytes.Buffer
			if err := link.Execute(&linkData, vars); err != nil {
				return err
			}
			if notification.WhatsApp.Document == nil {
				notification.WhatsApp.Document = &DocumentMessage{}
			}
			notification.WhatsApp.Document.Link = linkData.String()
			id, err := texttemplate.New(name).Funcs(f).Parse(n.Document.ID)
			if err != nil {
				return err
			}
			var idData bytes.Buffer
			if err := id.Execute(&idData, vars); err != nil {
				return err
			}
			notification.WhatsApp.Document.ID = idData.String()
		case "template":
			languageCode, err := texttemplate.New(name).Funcs(f).Parse(n.Template.Language.Code)
			if err != nil {
				return err
			}
			var languageCodeData bytes.Buffer
			if err := languageCode.Execute(&languageCodeData, vars); err != nil {
				return err
			}
			if notification.WhatsApp.Template == nil {
				notification.WhatsApp.Template = &TemplateMessage{}
			}
			notification.WhatsApp.Template.Language.Code = languageCodeData.String()
		}
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
	SubType    string              `json:"sub_type,omitempty"` // for buttons
	Index      string              `json:"index,omitempty"`    // for buttons
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

func NewButtonMessage(to, body string, buttons []InteractiveButton) *WhatsAppNotification {
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
				Buttons: buttons,
			},
		},
	}
}

func NewTemplate(
	to, name, lang string,
	components []TemplateComponent,
) *WhatsAppNotification {
	return &WhatsAppNotification{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             TypeTemplate,
		Template: &TemplateMessage{
			Name: name,
			Language: TemplateLanguage{
				Code: lang,
			},
			Components: components,
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
	PhoneNumberID      string `json:"phoneNumberID"`
	Token              string `json:"token"`
	ApiURL             string `json:"apiURL"`
	ApiVersion         string `json:"apiVersion"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	httputil.TransportOptions
}

type whatsappService struct {
	opts WhatsAppOptions
}

func NewWhatsAppService(opts WhatsAppOptions) NotificationService {
	return &whatsappService{opts: opts}
}

func (w *whatsappService) Send(notification Notification, dest Destination) (err error) {
	client, err := httputil.NewServiceHTTPClient(w.opts.TransportOptions, w.opts.InsecureSkipVerify, w.opts.ApiURL, "whatsapp")
	if err != nil {
		return err
	}
	var wn *WhatsAppNotification
	recipient := strings.TrimLeft(dest.Recipient, "+")
	if notification.WhatsApp == nil {
		wn = NewTextMessage(recipient, notification.Message, true)
	} else {
		switch notification.WhatsApp.Type {
		case "text":
			wn = NewTextMessage(recipient, notification.Message, notification.WhatsApp.Text.PreviewURL)
		case "image":
			wn = NewImageMessageByURL(recipient, notification.WhatsApp.Image.Link, notification.Message)
		case "document":
			wn = NewDocumentMessageByURL(recipient, notification.WhatsApp.Document.Link, notification.WhatsApp.Document.Filename, notification.Message)
		case "interactive":
			wn = NewButtonMessage(recipient, notification.Message, notification.WhatsApp.Interactive.Action.Buttons)
		case "template":
			wn = NewTemplate(recipient, notification.Message, notification.WhatsApp.Template.Language.Code, notification.WhatsApp.Template.Components)
		default:
			return fmt.Errorf("unknown type: %s", notification.WhatsApp.Type)
		}
	}

	b, _ := json.Marshal(wn)

	uri := fmt.Sprintf("/%s/%s/messages", w.opts.ApiVersion, w.opts.PhoneNumberID)
	req, err := http.NewRequest(http.MethodPost, w.opts.ApiURL+uri, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.opts.Token)

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
