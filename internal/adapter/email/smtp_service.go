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

// SMTPService implements port.EmailService over net/smtp with STARTTLS.
type SMTPService struct {
	cfg       Config
	templates *template.Template
}

// compile-time assertion that SMTPService satisfies the port interface.
var _ port.EmailService = (*SMTPService)(nil)

// NewSMTPService creates a new SMTPService and parses the email templates.
func NewSMTPService(cfg Config) (*SMTPService, error) {
	tmpl, err := template.New("email").Parse("")
	if err != nil {
		return nil, err
	}
	for name, body := range templateBodies {
		if _, err := tmpl.New(name).Parse(body); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
	}
	return &SMTPService{cfg: cfg, templates: tmpl}, nil
}

// templateData is the value passed to every email template.
type templateData struct {
	Shift        *domain.Shift
	Event        *domain.Event
	Registration *domain.Registration
	Member       *domain.Member
	Year         *domain.ClubYear
	ConfirmURL   string
	DaysUntil    int
	MissingHours float64
}

const (
	tmplKioskConfirmation = "kiosk-confirmation"
	tmplShiftConfirmation = "shift-confirmation"
	tmplShiftCancellation = "shift-cancellation"
	tmplReminder          = "reminder"
	tmplMissingHours      = "missing-hours"
)

var templateBodies = map[string]string{
	tmplKioskConfirmation: `Hallo,

vielen Dank fuer deine Anmeldung zur Schicht "{{.Shift.Name}}" bei der Veranstaltung "{{.Event.Name}}".

Bitte bestaetige deine Anmeldung ueber folgenden Link:
{{.ConfirmURL}}

Schichtbeginn: {{.Shift.StartAt.Format "02.01.2006 15:04"}} Uhr

Viele Gruesse
TSC Schwarz-Gelb Aachen`,

	tmplShiftConfirmation: `Hallo,

deine Anmeldung zur Schicht "{{.Shift.Name}}" bei "{{.Event.Name}}" ist bestaetigt.

Ort: {{.Event.Location}}
Beginn: {{.Shift.StartAt.Format "02.01.2006 15:04"}} Uhr
Ende: {{.Shift.EndAt.Format "02.01.2006 15:04"}} Uhr

Viele Gruesse
TSC Schwarz-Gelb Aachen`,

	tmplShiftCancellation: `Hallo,

deine Anmeldung zur Schicht "{{.Shift.Name}}" bei "{{.Event.Name}}" wurde storniert.

Falls dies ein Versehen war, kannst du dich erneut anmelden.

Viele Gruesse
TSC Schwarz-Gelb Aachen`,

	tmplReminder: `Hallo,

dies ist eine Erinnerung an deine Schicht "{{.Shift.Name}}" bei "{{.Event.Name}}".

Die Schicht beginnt in {{.DaysUntil}} Tag(en) am {{.Shift.StartAt.Format "02.01.2006 15:04"}} Uhr.
Ort: {{.Event.Location}}

Viele Gruesse
TSC Schwarz-Gelb Aachen`,

	tmplMissingHours: `Hallo {{.Member.FirstName}},

du hast im Vereinsjahr "{{.Year.Label}}" noch {{printf "%.1f" .MissingHours}} offene Stunden.

Bitte melde dich rechtzeitig fuer weitere Schichten an, um deine Stunden zu erfuellen.

Viele Gruesse
TSC Schwarz-Gelb Aachen`,
}

// SendConfirmation sends a registration confirmation email.
func (s *SMTPService) SendConfirmation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	return s.send(ctx, to, subjectFor(event, "Anmeldung bestaetigt"), tmplShiftConfirmation, templateData{
		Registration: reg, Shift: shift, Event: event,
	})
}

// SendReminder sends a shift reminder email.
func (s *SMTPService) SendReminder(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event, daysUntil int) error {
	return s.send(ctx, to, subjectFor(event, "Erinnerung an deine Schicht"), tmplReminder, templateData{
		Registration: reg, Shift: shift, Event: event, DaysUntil: daysUntil,
	})
}

// SendCancellation sends a cancellation notification email.
func (s *SMTPService) SendCancellation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	return s.send(ctx, to, subjectFor(event, "Anmeldung storniert"), tmplShiftCancellation, templateData{
		Registration: reg, Shift: shift, Event: event,
	})
}

// SendKioskConfirmation sends a kiosk double-opt-in confirmation email.
func (s *SMTPService) SendKioskConfirmation(ctx context.Context, to string, confirmURL string, shift *domain.Shift, event *domain.Event) error {
	return s.send(ctx, to, subjectFor(event, "Bitte bestaetige deine Anmeldung"), tmplKioskConfirmation, templateData{
		Shift: shift, Event: event, ConfirmURL: confirmURL,
	})
}

// SendMissingHoursWarning sends a warning about unfulfilled hours.
func (s *SMTPService) SendMissingHoursWarning(ctx context.Context, to string, member *domain.Member, missingHours float64, year *domain.ClubYear) error {
	return s.send(ctx, to, "Offene Stunden im Vereinsjahr", tmplMissingHours, templateData{
		Member: member, MissingHours: missingHours, Year: year,
	})
}

func subjectFor(event *domain.Event, prefix string) string {
	if event != nil && event.Name != "" {
		return fmt.Sprintf("%s: %s", prefix, event.Name)
	}
	return prefix
}

// send renders the named template and delivers the message via SMTP+STARTTLS.
func (s *SMTPService) send(ctx context.Context, to, subject, templateName string, data templateData) error {
	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("render email template %s: %w", templateName, err)
	}

	msg := s.buildMessage(to, subject, body.String())
	return s.deliver(ctx, to, msg)
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
