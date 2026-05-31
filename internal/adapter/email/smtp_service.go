package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// Config holds the SMTP connection settings for the email service.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string
}

// templateStore is the subset of port.EmailTemplateRepository the email service
// needs to look up overridable templates. It is optional; when nil the built-in
// defaults are always used.
type templateStore interface {
	GetTemplate(ctx context.Context, name string) (*domain.EmailTemplate, error)
}

// SMTPService implements port.EmailService over net/smtp with STARTTLS. It renders
// email bodies from templates stored in the database (Section 4), falling back to
// built-in defaults when a template row is missing.
type SMTPService struct {
	cfg   Config
	store templateStore
	log   port.EmailLogRepository // optional; logs every send attempt (N-004)
}

// compile-time assertion that SMTPService satisfies the port interface.
var _ port.EmailService = (*SMTPService)(nil)

// NewSMTPService creates a new SMTPService. store may be nil to always use the
// built-in default templates; log may be nil to disable send logging.
func NewSMTPService(cfg Config, store templateStore, log port.EmailLogRepository) (*SMTPService, error) {
	return &SMTPService{cfg: cfg, store: store, log: log}, nil
}

// defaultTemplates are the built-in fallback subject/body pairs keyed by template
// name. They use the same flattened placeholders as the seeded DB rows.
var defaultTemplates = map[string]struct{ Subject, Body string }{
	domain.EmailTemplateKioskConfirmation: {
		"Bitte bestaetige deine Anmeldung: {{.EventName}}",
		"Hallo {{.MemberName}},\n\nvielen Dank fuer deine Anmeldung zur Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\".\nBitte bestaetige ueber:\n{{.ConfirmURL}}\n\nBeginn: {{.StartAt}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateShiftConfirmation: {
		"Anmeldung bestaetigt: {{.EventName}}",
		"Hallo {{.MemberName}},\n\ndeine Anmeldung zur Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\" ist bestaetigt.\nOrt: {{.Location}}\nBeginn: {{.StartAt}}\nEnde: {{.EndAt}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateShiftDeregister: {
		"Abmeldung bestaetigt: {{.EventName}}",
		"Hallo {{.MemberName}},\n\ndeine Abmeldung von \"{{.ShiftName}}\" bei \"{{.EventName}}\" wurde verarbeitet.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateReminder1W: {
		"Erinnerung: deine Schicht in einer Woche",
		"Hallo {{.MemberName}},\n\nErinnerung an deine Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\" in {{.DaysUntil}} Tag(en) am {{.StartAt}}.\nOrt: {{.Location}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateReminder1D: {
		"Erinnerung: deine Schicht morgen",
		"Hallo {{.MemberName}},\n\nErinnerung an deine Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\" in {{.DaysUntil}} Tag(en) am {{.StartAt}}.\nOrt: {{.Location}}\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateShiftCancelled: {
		"Schicht abgesagt: {{.EventName}}",
		"Hallo {{.MemberName}},\n\ndie Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\" wurde abgesagt.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateYearBilling: {
		"Beitragsabrechnung Vereinsjahr {{.YearLabel}}",
		"Hallo {{.MemberName}},\n\nim Vereinsjahr \"{{.YearLabel}}\" fehlen {{.MissingHours}} Stunden. Beitrag: {{.AmountEUR}} EUR.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateMissingHours: {
		"Offene Stunden im Vereinsjahr {{.YearLabel}}",
		"Hallo {{.MemberName}},\n\ndu hast im Vereinsjahr \"{{.YearLabel}}\" noch {{.MissingHours}} offene Stunden.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
	domain.EmailTemplateUnderstaffed: {
		"Schicht unterbesetzt: {{.EventName}}",
		"Hallo,\n\ndie Schicht \"{{.ShiftName}}\" bei \"{{.EventName}}\" am {{.StartAt}} ist unterbesetzt.\n\nViele Gruesse\nTSC Schwarz-Gelb Aachen",
	},
}

const dateLayout = "02.01.2006 15:04"

// shiftData builds the flattened placeholder map for shift/event emails.
func shiftData(member *domain.Member, shift *domain.Shift, event *domain.Event, extra map[string]any) map[string]any {
	d := map[string]any{
		"MemberName": "",
		"ShiftName":  "",
		"EventName":  "",
		"Location":   "",
		"StartAt":    "",
		"EndAt":      "",
		"ConfirmURL": "",
		"DaysUntil":  0,
	}
	if member != nil {
		d["MemberName"] = member.FirstName
	}
	if shift != nil {
		d["ShiftName"] = shift.Name
		d["StartAt"] = shift.StartAt.Format(dateLayout)
		d["EndAt"] = shift.EndAt.Format(dateLayout)
	}
	if event != nil {
		d["EventName"] = event.Name
		d["Location"] = event.Location
	}
	for k, v := range extra {
		d[k] = v
	}
	return d
}

// SendConfirmation sends a registration confirmation email.
func (s *SMTPService) SendConfirmation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	return s.sendTemplate(ctx, to, domain.EmailTemplateShiftConfirmation, shiftData(nil, shift, event, nil))
}

// SendReminder sends a shift reminder email. The one-week reminder uses the
// reminder-1w template, all other lead times use reminder-1d.
func (s *SMTPService) SendReminder(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event, daysUntil int) error {
	name := domain.EmailTemplateReminder1D
	if daysUntil >= 7 {
		name = domain.EmailTemplateReminder1W
	}
	return s.sendTemplate(ctx, to, name, shiftData(nil, shift, event, map[string]any{"DaysUntil": daysUntil}))
}

// SendCancellation sends a deregistration notification email.
func (s *SMTPService) SendCancellation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	return s.sendTemplate(ctx, to, domain.EmailTemplateShiftDeregister, shiftData(nil, shift, event, nil))
}

// SendKioskConfirmation sends a kiosk double-opt-in confirmation email.
func (s *SMTPService) SendKioskConfirmation(ctx context.Context, to string, confirmURL string, shift *domain.Shift, event *domain.Event) error {
	return s.sendTemplate(ctx, to, domain.EmailTemplateKioskConfirmation, shiftData(nil, shift, event, map[string]any{"ConfirmURL": confirmURL}))
}

// SendMissingHoursWarning sends a warning about unfulfilled hours.
func (s *SMTPService) SendMissingHoursWarning(ctx context.Context, to string, member *domain.Member, missingHours float64, year *domain.ClubYear) error {
	data := map[string]any{"MissingHours": fmt.Sprintf("%.1f", missingHours), "YearLabel": ""}
	if member != nil {
		data["MemberName"] = member.FirstName
	}
	if year != nil {
		data["YearLabel"] = year.Label
	}
	return s.sendTemplate(ctx, to, domain.EmailTemplateMissingHours, data)
}

// SendYearBilling sends the year-end billing email.
func (s *SMTPService) SendYearBilling(ctx context.Context, to string, member *domain.Member, missingHours float64, amountCents int, year *domain.ClubYear) error {
	data := map[string]any{
		"MissingHours": fmt.Sprintf("%.1f", missingHours),
		"AmountEUR":    fmt.Sprintf("%.2f", float64(amountCents)/100.0),
		"YearLabel":    "",
	}
	if member != nil {
		data["MemberName"] = member.FirstName
	}
	if year != nil {
		data["YearLabel"] = year.Label
	}
	return s.sendTemplate(ctx, to, domain.EmailTemplateYearBilling, data)
}

// SendUnderstaffedNotice notifies an organizer that a shift is understaffed.
func (s *SMTPService) SendUnderstaffedNotice(ctx context.Context, to string, shift *domain.Shift, event *domain.Event) error {
	return s.sendTemplate(ctx, to, domain.EmailTemplateUnderstaffed, shiftData(nil, shift, event, nil))
}

// SendByTemplate renders an arbitrary stored template by name and sends it.
func (s *SMTPService) SendByTemplate(ctx context.Context, to, templateName string, data map[string]any) error {
	return s.sendTemplate(ctx, to, templateName, data)
}

// resolveTemplate returns the subject and body sources for a template name,
// preferring a stored DB row and falling back to the built-in default.
func (s *SMTPService) resolveTemplate(ctx context.Context, name string) (subject, body string) {
	if s.store != nil {
		if t, err := s.store.GetTemplate(ctx, name); err == nil && t != nil {
			return t.Subject, t.Body
		}
	}
	if d, ok := defaultTemplates[name]; ok {
		return d.Subject, d.Body
	}
	return name, ""
}

// sendTemplate resolves, renders, delivers and logs a single email.
func (s *SMTPService) sendTemplate(ctx context.Context, to, templateName string, data map[string]any) error {
	subjectSrc, bodySrc := s.resolveTemplate(ctx, templateName)

	subject, err := renderString("subject:"+templateName, subjectSrc, data)
	if err != nil {
		return s.record(ctx, to, templateName, subjectSrc, "", err)
	}
	body, err := renderString("body:"+templateName, bodySrc, data)
	if err != nil {
		return s.record(ctx, to, templateName, subject, "", err)
	}

	msg := s.buildMessage(to, subject, body)
	deliverErr := s.deliver(ctx, to, msg)
	return s.record(ctx, to, templateName, subject, body, deliverErr)
}

// record logs the send attempt (when a log repo is configured) and returns err.
func (s *SMTPService) record(ctx context.Context, to, templateName, subject, body string, err error) error {
	if s.log != nil {
		entry := &domain.EmailLogEntry{
			To:        to,
			Template:  templateName,
			Subject:   subject,
			Body:      body,
			Status:    domain.EmailLogStatusSent,
			CreatedAt: time.Now().UTC(),
		}
		if err != nil {
			entry.Status = domain.EmailLogStatusFailed
			entry.Error = err.Error()
		}
		_ = s.log.Insert(ctx, entry)
	}
	return err
}

// renderString renders a text/template source against data.
func renderString(name, src string, data map[string]any) (string, error) {
	t, err := template.New(name).Parse(src)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", name, err)
	}
	return buf.String(), nil
}

func (s *SMTPService) buildMessage(to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + s.cfg.From + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

// deliver opens an SMTP connection, upgrades to TLS via STARTTLS when offered,
// authenticates if credentials are configured, and sends the message.
func (s *SMTPService) deliver(ctx context.Context, to string, msg []byte) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("new smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if s.cfg.Username != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		_ = wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}
