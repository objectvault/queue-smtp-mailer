package mailer

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"errors"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"

	mailyak "github.com/domodwyer/mailyak/v3"

	"github.com/objectvault/queue-interface/messages"
	"github.com/objectvault/queue-smtp-mailer/config"
)

var _connection string
var _auth smtp.Auth

func getSMTPConnection(c *config.DaemonConfig) string {
	// Have we already setup a SMTP connection?
	if _connection == "" { // NO: Setup
		// Setup Connection Information
		host := c.SMTPRelay.Server.Host
		port := c.SMTPRelay.Server.Port
		if port == 0 {
			_connection = host + ":" + "25"
		} else {
			_connection = host + ":" + strconv.Itoa(port)
		}
	}

	return _connection
}

func getSMTPAuthentication(c *config.DaemonConfig) smtp.Auth {
	// Have we already setup a SMTP connection?
	if _auth == nil { // NO: Setup
		if c.SMTPRelay.Authentication != nil {
			user := c.SMTPRelay.Authentication.User
			if user != "" {
				_auth = smtp.PlainAuth("",
					user,
					c.SMTPRelay.Authentication.Password,
					c.SMTPRelay.Server.Host)
			}
		}
	}

	return _auth
}

func templatePath(c *config.DaemonConfig, name string, t string) string {
	base := c.Paths.Templates
	path := filepath.Join(base, name+"."+t+".template")

	// Does Template File Exist
	i, err := os.Stat(path)
	if os.IsNotExist(err) || i.IsDir() { // NO
		return ""
	}
	// ELSE: YES
	return path
}

func SendMail(c *config.DaemonConfig, msg *messages.EmailMessage) error {
	template := msg.Template()
	log.Printf("Email Template [%s]", template)

	// Create a new email - specify the SMTP host and auth
	email := mailyak.New(getSMTPConnection(c), getSMTPAuthentication(c))

	// Initialize Basics
	email.To(msg.To())
	email.From(msg.From("noreply@test-to.com"))
	email.FromName("Do Not Reply")

	email.Subject("User Activation")

	// Mail Template Files Template
	textTemplate := templatePath(c, template, "text")
	htmlTemplate := templatePath(c, template, "html")

	// Does Template Exist?
	if (textTemplate == "") && (htmlTemplate == "") { // NO
		return errors.New("Invalid Template [" + template + "]")
	}

	if textTemplate != "" {
		expandTextTemplate(textTemplate, msg.GetParameters(), email.Plain())
	}

	if htmlTemplate != "" {
		expandHTMLTemplate(htmlTemplate, msg.GetParameters(), email.HTML())
	}

	// Send Email
	err := email.Send()
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
