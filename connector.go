package main

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

	"github.com/objectvault/queue-smtp-mailer/config"
	"github.com/objectvault/queue-smtp-mailer/poller"
)

func connector(c *config.DaemonConfig) {
	// After everything is Done Make Sure to Close Everything
	defer func() {
		if poller.Shutdown {
			done <- true
		}
	}()

	// Control Settings
	maxRetries := c.Options.ConnectionRetriesMax
	interval := time.Duration(c.Options.ConnectionRetryInterval*1000) * time.Millisecond

	// Number of Sequential Errors
	errorCount := 0

	log.Print("START: Connection Poller")
	log.Printf("Max Retries [%d]", maxRetries)
	log.Printf("POLL Interval [%d]s", c.Options.ConnectionRetryInterval)
	log.Print("START: Connection Poller")

	// ENDLESS Loop
	for {
		// Log Retry
		log.Printf("Retry [%d] of [%d]", errorCount+1, maxRetries)

		// Do we Want to Stop the Poller?
		if poller.Shutdown { // YES: Break Out of Loop
			log.Print("Stopping Connection Poller...")
			break
		}

		// Increment Error Counter
		errorCount++

		// Do we have a Connection?
		_, err := mailerMQ.OpenConnection()
		if err == nil { // YES: Start Message Poller
			errorCount = 0 // Reset Error Count
			poller.Poller(c, mailerMQ)

			// Poller Stopped - Presume Bad Connection - Reset it
			mailerMQ.CloseConnection()
		}

		// Did we exceed retry count?
		if (maxRetries > 0) && (errorCount > maxRetries) { // YES: Shutdown the Server
			poller.Shutdown = true
			break
		}
		// ELSE: NO

		// Sleep and Retry Connection
		log.Print("Sleeping...")
		time.Sleep(interval)
	}

	log.Print("STOP: Connection Poller")
}
