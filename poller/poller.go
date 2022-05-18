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
	"log"
	"time"

	"github.com/objectvault/queue-interface/queue"
	"github.com/objectvault/queue-smtp-mailer/config"
)

// Flags
var Shutdown bool // Shutdown Poller

func Poller(c *config.DaemonConfig, mailerMQ *queue.AMQPServerConnection) {
	// Number of Sequential Errors
	errorCount := 0

	// Poller Defaults
	maxMessages := c.Options.PollMaxMessages
	interval := time.Duration(c.Options.PollInterval*1000) * time.Millisecond
	name := "inbox"

	log.Print("START: Message Poller")
	log.Printf("Max of Messages [%d] per POLL", maxMessages)
	log.Printf("POLL Interval [%d]s", c.Options.PollInterval)
	log.Printf("POLL Queue [%s]", name)

	// ENDLESS Loop
	for {
		// Do we Want to Stop the Poller?
		if Shutdown { // YES: Break Out of Loop
			log.Print("Stopping Message Poller...")
			break
		}

		log.Print("Retrieving Messages...")
		for i := 0; i < maxMessages; i++ {
			delivery, err := mailerMQ.QueueRetrieve("read", name)

			if err != nil {
				log.Printf("Error [%d] Reading Message...", errorCount)

				errorCount++
				if errorCount > 10 {
					log.Print("Too Many Error. Stopping Poller...")
					log.Print("STOP: Message Poller")
					return
				}

				continue
			}

			// Reset Error Count
			errorCount = 0

			// Is Queue Empty?
			if delivery == nil { // YES
				log.Print("Queue Empty. Sleeping...")
				break
			}

			// Start Mailer Thread
			go process(c, delivery)
		}

		log.Print("Sleeping...")
		time.Sleep(interval)
	}

	log.Print("STOP: Message Poller")
}
