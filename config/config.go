package config

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
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/objectvault/queue-interface/shared"
)

type Paths struct {
	Templates string `json:"templates"`        // Templates Directory
	Output    string `json:"output,omitempty"` // Message Output Directory
	Temporary string `json:"tmp,omitempty"`    // Temporary Directory
}

type Authentication struct {
	User     string `json:"user,omitempty"`     // User Name
	Password string `json:"password,omitempty"` // User Password
}

type SMTPRelay struct {
	Server         *shared.Server  `json:"server,omitempty"`         // Email Relay Server
	Authentication *Authentication `json:"authentication,omitempty"` // Email Relay Server
}

type Retries struct {
	RetriesMax      int `json:"max-retries,omitempty"`    // Limit of Retry Attempts (0 - No Limit)
	RetriesInterval int `json:"retry-interval,omitempty"` // Seconds Between Retries (DEFAULT 60 seconds)
}

type Options struct {
	ConnectionRetriesMax    int    `json:"conn-max-retries,omitempty"`    // Limit of Retry Attempts (0 - No Limit)
	ConnectionRetryInterval int    `json:"conn-retry-interval,omitempty"` // Seconds Between Retries (DEFAULT 60 seconds)
	PollMaxMessages         int    `json:"poll-max-messages,omitempty"`   // Maximum Messages Processed per Poll (DEFAULT 10 seconds)
	PollInterval            int    `json:"poll-interval,omitempty"`       // Seconds Between Poll (DEFAULT 10 seconds)
	PollQueue               string `json:"poll-queue,omitempty"`          // Name of Incoming Queue
}

type DaemonConfig struct {
	Queue     *shared.Queue `json:"queue,omitempty"`   // List of AMQP Servers
	SMTPRelay *SMTPRelay    `json:"relay,omitempty"`   // Email Relay Server
	Paths     *Paths        `json:"paths,omitempty"`   // Paths to Use
	Options   *Options      `json:"options,omitempty"` // Server Options
}

// Config CONTAINER for Daemon CONFIGURATION
var serverConfig *DaemonConfig

func Config() *DaemonConfig {
	return serverConfig
}

// Load Configuration File
func Load(path string) (*DaemonConfig, error) {
	var config DaemonConfig

	// Open Configuration File
	file, errFile := os.Open(path)
	if errFile != nil {
		log.Printf("Error [%s]", errFile)
		return nil, errors.New("ERROR: Configuration File Required")
	}

	// Close File After (Try) Loading
	defer file.Close()

	// Decode JSON
	decoder := json.NewDecoder(file)
	errDecoder := decoder.Decode(&config)
	if errDecoder != nil {
		log.Printf("JSON Parse Error [%s]\n", errDecoder)
		return nil, errors.New("ERROR: Invalid Configuration File")
	}

	// Do we have AMQP Host Addresses?
	if (config.Queue == nil) || len(config.Queue.Servers) == 0 { // NO: Abort
		log.Print("No Queue Server Connection Information")
		return nil, errors.New("ERROR: Invalid Configuration File")
	}

	// Do we have SMTP Relay Configuration?
	if (config.SMTPRelay == nil) || (config.SMTPRelay.Server == nil) { // NO: Abort
		log.Print("No SMTP Relay Connection Information")
		return nil, errors.New("ERROR: Invalid Configuration File")
	}

	// Do we have SMTP Authentication?
	if config.SMTPRelay.Authentication == nil { // NO: Warn
		log.Print("WARNING: No SMTP Authentication")
	}

	// Configuration Path Exists?
	if config.Paths == nil { // NO: Create Default
		config.Paths = &Paths{
			Templates: "./templates",
		}
	}

	// Template Directory Provided?
	if config.Paths.Templates == "" { // NO
		log.Print("No Template Directory Specified")
		return nil, errors.New("ERROR: Invalid Configuration File")
	}

	// Does Template Directory Exist?
	i, err := os.Stat(config.Paths.Templates)
	if os.IsNotExist(err) || !i.IsDir() { // NO
		log.Printf("[%s] Does not Exist or is Not a Directory", config.Paths.Templates)
		return nil, errors.New("ERROR: Invalid Configuration File")
	}

	// Any Options Set?
	if config.Options == nil { // NO: Need at least a Queue
		log.Print("No Message Queue Name set in Configuration File")
		return nil, errors.New("ERROR: Invalid Configuration File")
	} else {
		// Do we have a Valid Queue Name?
		if config.Options.PollQueue == "" { // NO: Abort
			log.Print("No Message Queue Name set in Configuration File")
			return nil, errors.New("ERROR: Invalid Configuration File")
		}

		// Do we have a Valid Retry Interval?
		if config.Options.ConnectionRetryInterval <= 0 { // NO: Set Default 60 seconds
			config.Options.ConnectionRetryInterval = 60
		}

		// Does the Poller Have a Max Messages Limit?
		if config.Options.PollMaxMessages <= 1 { // NO: Set Default 10 Messages
			config.Options.PollMaxMessages = 10
		}

		// Do we have Wait Interval?
		if config.Options.PollInterval <= 0 { // NO: Set Default 10 seconds
			config.Options.PollInterval = 10
		}
	}

	// Convert Path to Full Path Name
	config.Paths.Templates, _ = filepath.Abs(config.Paths.Templates)
	log.Printf("TEMPLATE DIR [%s]", config.Paths.Templates)

	// Return Configuration
	return &config, nil
}

func getChildProperty(source map[string]interface{}, elements []string, i int, dvalue interface{}) interface{} {
	if i >= len(elements) {
		return source
	}

	element := strings.TrimSpace(elements[i])
	if len(element) > 0 {
		value, exists := source[elements[i]]
		if exists {
			switch v := value.(type) {
			case map[string]interface{}:
				return getChildProperty(v, elements, i+1, dvalue)
			default:
				return v
			}
		}
	}

	return dvalue
}

func ConfigProperty(source map[string]interface{}, path string, dvalue interface{}) interface{} {
	elements := strings.Split(path, ".")
	if source == nil {
		return dvalue
	}

	element := strings.TrimSpace(elements[0])
	if len(element) > 0 {
		value, exists := source[elements[0]]
		if exists {
			switch v := value.(type) {
			case map[string]interface{}:
				return getChildProperty(v, elements, 1, dvalue)
			default:
				return v
			}
		}
	}

	return dvalue
}
