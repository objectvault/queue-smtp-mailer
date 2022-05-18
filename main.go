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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/objectvault/queue-interface/queue"
	"github.com/objectvault/queue-interface/shared"
	"github.com/objectvault/queue-smtp-mailer/config"
	"github.com/objectvault/queue-smtp-mailer/poller"
)

// Channels to Control Daemon
var signals chan os.Signal
var done chan bool

// Queue Connection
var mailerMQ *queue.AMQPServerConnection

func setMQConnection(c *shared.Queue) (*queue.AMQPServerConnection, error) {
	q := &queue.AMQPServerConnection{}

	if c == nil {
		return nil, errors.New("[setMQConnection] No Configuration for Queue")
	}

	q.SetConnection(c.Servers)
	q.SetPrefix(c.QueuePrefix)
	return q, nil
}

// MAIN //
func main() {
	// COMMAND LINE PARSER //
	flag.Usage = func() {
		usage := `
		AMQP Queue Mailer

		Usage:
		  server -c /path/to/conf
		  server -v | --version
		  server -h | --help

		  Options:
		    -h --help     Show this screen.
		    -v            Show version.
		    -c            Path to configuration file [default: ./mailer.json].
		`

		fmt.Println(usage)
	}
	sConfPath := flag.String("c", "./mailer.json", "Path to configuration file")
	bVersion := flag.Bool("v", false, "Path to configuration file")
	flag.Parse()

	// Version Flag Set?
	if *bVersion { // YES: Display Version and Exit
		fmt.Println("AMQP Queue Mailer [0.0.1]")
		os.Exit(0)
	}

	// LOG
	log.Print("Starting Daemon")

	// Load Configuration File
	log.Print("Loading Configuration File")
	c, err := config.Load(*sConfPath)
	if err != nil {
		log.Fatal(err)
	}

	// Set Message Queue Connection Settings
	mailerMQ, _ = setMQConnection(c.Queue)

	// After everything is Done Make Sure to Close Everything
	defer func() {
		log.Print("EXITING: Close All Connections")

		// Queue Connection Established?
		if mailerMQ != nil { // YES: Close it
			mailerMQ.CloseConnection()
		}
	}()

	// Clear Shutdown Flag
	poller.Shutdown = false

	// Create Channels
	signals = make(chan os.Signal, 1)
	done = make(chan bool, 1)

	// Capture Termination Signals
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// Create Signal Handler
	go func() {
		// Capture Signals
		s := <-signals

		// Output
		fmt.Println()
		fmt.Println(s)

		// Stop Daemon
		log.Print("Starting Shutdown Process")

		// Stop Other Threads
		poller.Shutdown = true
	}()

	// Start Connection Thread
	go connector(c)

	// Capture Done Signal
	<-done
	log.Print("Exiting Daemon")
}
