package alerts

import (
	log "github.com/Sirupsen/logrus"
	constants "github.com/megamsys/libgo/utils"
	"net/smtp"
	"strings"
)

const (
	LAUNCHED EventAction = iota
	DESTROYED
	STATUS
	DEDUCT
	ONBOARD
	RESET
	INVITE
	BALANCE
	INVOICE
	BILLEDHISTORY
	TRANSACTION
	DESCRIPTION
	SNAPSHOTTING
	SNAPSHOTTED
	RUNNING
	FAILURE
	INSUFFICIENT_FUND
	QUOTA_UNPAID
	SKEWS_ACTIONS
	SKEWS_WARNING
)

var Mailer map[string]string

type Notifier interface {
	Notify(eva EventAction, edata EventData) error
	satisfied(eva EventAction) bool
}

// Extra information about an event.
type EventData struct {
	M map[string]string
	D []string
}

type EventAction int

func (v *EventAction) String() string {
	switch *v {
	case LAUNCHED:
		return "launched"
	case DESTROYED:
		return "destroyed"
	case STATUS:
		return "status"
	case DEDUCT:
		return "deduct"
	case ONBOARD:
		return "onboard"
	case RESET:
		return "reset"
	case INVITE:
		return "invite"
	case BALANCE:
		return "balance"
	case INSUFFICIENT_FUND:
		return "insufficientfunds"
	case QUOTA_UNPAID:
		return "quotaoverdue"
	case DESCRIPTION:
		return "description"
	case SNAPSHOTTING:
		return "snapshotting"
	case SNAPSHOTTED:
		return "snapshotted"
	case RUNNING:
		return "running"
	case FAILURE:
		return "failure"
	case SKEWS_WARNING:
		return constants.SKEWS_WARNING
	default:
		return "arrgh"
	}
}

type mailer struct {
	username string
	password string
	identity string
	domain   string
	sender   string
	nilavu   string
	logo     string
	home     string
	dir      string
}

func NewMailer(m map[string]string, n map[string]string) Notifier {
	mg := &mailer{
		username: m[constants.USERNAME],
		password: m[constants.PASSWORD],
		identity: m[constants.IDENTITY],
		sender:   m[constants.SENDER],
		domain:   m[constants.DOMAIN],
		nilavu:   m[constants.NILAVU],
		logo:     m[constants.LOGO],
		home:     n[constants.HOME],
		dir:      n[constants.DIR],
	}
	mg.makeGlobal()
	return mg
}

func (m *mailer) makeGlobal() {
	mm := make(map[string]string, 0)
	mm[constants.USERNAME] = m.username
	mm[constants.PASSWORD] = m.password
	mm[constants.IDENTITY] = m.identity
	mm[constants.SENDER] = m.sender
	mm[constants.DOMAIN] = m.domain
	mm[constants.NILAVU] = m.nilavu
	mm[constants.LOGO] = m.logo
	mm[constants.HOME] = m.home
	mm[constants.DIR] = m.dir
	Mailer = mm
}

func (m *mailer) satisfied(eva EventAction) bool {
	if eva == STATUS {
		return false
	}
	return true
}

func (m *mailer) Notify(eva EventAction, edata EventData) error {
	if !m.satisfied(eva) {
		return nil
	}
	edata.M[constants.NILAVU] = m.nilavu
	edata.M[constants.LOGO] = m.logo

	bdy, err := body(eva.String(), edata.M, m.dir)
	if err != nil {
		return err
	}

	return m.Send(bdy, "", subject(eva), edata.M[constants.EMAIL])
}

func (m *mailer) Send(bdy string, sender string, subject string, receiver string) error {
	var addr string
	if len(strings.TrimSpace(sender)) <= 0 {
		sender = m.sender
	}

	auth := smtp.PlainAuth(m.identity, m.username, m.password, m.domain)

	msg := "From: " + sender + "\r\n" +
		"To: " + receiver + "\r\n" +
		"MIME-Version: 1.0" + "\r\n" +
		"Content-type: text/html" + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		bdy + "\r\n"

	domain := strings.Split(m.domain, ":")
	if len(domain) < 2 || domain[1] == "" {
		addr = m.domain + ":587"
	} else {
		addr = m.domain
	}

	err := smtp.SendMail(addr, auth, sender, []string{receiver}, []byte(msg))
	if err != nil {
		return err
	}
	log.Infof("Mailgun sent to %v", receiver)
	return nil
}

func subject(eva EventAction) string {
	var sub string
	switch eva {
	case ONBOARD:
		sub = "Ahoy. Welcome aboard!"
	case RESET:
		sub = "You have fat finger.!"
	case INVITE:
		sub = "Lets party!"
	case BALANCE:
		sub = "Piggy bank!"
	case INSUFFICIENT_FUND:
		sub = "Insufficient funds!"
	case QUOTA_UNPAID:
		sub = "Payment pending!"
	case LAUNCHED:
		sub = "Up!"
	case RUNNING:
		sub = "Ahoy! Your application is running "
	case DESTROYED:
		sub = "Nuked"
	case SNAPSHOTTING:
		sub = "Snapshot creating!"
	case SNAPSHOTTED:
		sub = "Ahoy! Snapshot created"
	case FAILURE:
		sub = "Your application failure"
	case SKEWS_WARNING:
		return "Payment Reminder !"
	default:
		break
	}
	return sub
}
