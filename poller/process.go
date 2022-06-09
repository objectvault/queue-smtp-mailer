package poller

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
	"strings"

	"github.com/streadway/amqp"

	"github.com/objectvault/queue-interface/messages"
	"github.com/objectvault/queue-smtp-mailer/config"
	"github.com/objectvault/queue-smtp-mailer/mailer"
)

func extractEmailMesssage(msg *amqp.Delivery) (*messages.QueueMessage, error) {
	log.Printf("Processing Message [%s]", msg.MessageId)

	// Convert Delivery to Email Message
	queueMessage := messages.QueueMessage{}
	err := queueMessage.UnmarshalJSON(msg.Body)
	if err != nil {
		log.Printf("Error Decoding Email Request")
		return nil, err
	}

	return &queueMessage, nil
}

func setEmailMaps(message *messages.EmailMessage, s map[string]interface{}, d string) error {
	var err error
	for k, v := range s {
		// NOTE: We are using lower case keys to avoid case sensitive duplicates
		k = strings.ToLower(k)
		switch d {
		case "params":
			err = message.SetParameter(k, v.(string))
		case "headers":
			// Is Value a String?
			s, ok := v.(string)
			if !ok { // NO: Skip
				continue
			}
			err = message.SetHeader(k, s)
		}

		if err != nil {
			break
		}
	}

	return err
}

func toEmailMessage(source *map[string]interface{}) (*messages.EmailMessage, error) {

	// Create
	message := messages.EmailMessage{}

	params := make(map[string]interface{})

	// Range through the MAP and Set Message Properties
	var err error
	for k, v := range *source {
		k = strings.ToLower(k)
		switch k {
		case "template":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetTemplate(s)
			} else {
				err = errors.New("Invalid Value for 'template' field")
			}
		case "locale":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetLanguage(s)
			} else {
				err = errors.New("Invalid Value for 'locale' field")
			}
		case "to":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetTo(s)
			} else {
				err = errors.New("Invalid Value for 'to' field")
			}
		case "from":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetFrom(s)
			} else {
				err = errors.New("Invalid Value for 'from' field")
			}
		case "cc":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetCC(s)
			} else {
				err = errors.New("Invalid Value for 'cc' field")
			}
		case "bcc":
			s, castOK := v.(string)
			if castOK {
				_, err = message.SetBCC(s)
			} else {
				err = errors.New("Invalid Value for 'bcc' field")
			}
		case "params":
			m, castOK := v.(map[string]interface{})
			if castOK {
				err = setEmailMaps(&message, m, "params")
			} else {
				err = errors.New("Invalid Value for 'params' field")
			}
		case "headers":
			m, castOK := v.(map[string]interface{})
			if castOK {
				err = setEmailMaps(&message, m, "headers")
			} else {
				err = errors.New("Invalid Value for 'headers' field")
			}
		default: // Add Value to Parameters List
			s, castOK := v.(string)
			if castOK {
				params[k] = s
			}
		}

		if err != nil {
			return nil, err
		}
	}

	err = setEmailMaps(&message, params, "params")
	if err != nil {
		return nil, err
	}

	if !message.IsValid() {
		return nil, errors.New("Invalid Email Message Request")
	}

	return &message, nil
}

func process(c *config.DaemonConfig, d *amqp.Delivery) error {
	// STEP 1: Extract Queue Message //
	msg, err := extractEmailMesssage(d)
	if err != nil {
		log.Print("Queue Message is Invalid")
	}

	log.Printf("Processing Message [%s]", msg.ID())

	// STEP 2: Extract Email Request //
	i := msg.Message()

	// Is Valid Message Format?
	s, ok := (*i).(map[string]interface{})
	if !ok { // NO
		return errors.New("Invalid Massage Format")
	}

	// Import Message Date into Object
	emailMessage, err := toEmailMessage(&s)
	if err != nil {
		log.Print(err)
		return err
	}

	// STEP 3: Try to Send Email
	err = mailer.SendMail(c, emailMessage)
	if err != nil {
		log.Print(err)
		return err
	}

	// STEP 4: Acknowledge Message so it's removed from Queue
	err = d.Ack(false)
	if err != nil {
		log.Print(err)
	}

	return nil
}
