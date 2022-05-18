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
	"io"
	"log"

	"text/template"
)

func expandTextTemplate(path string, params interface{}, w io.Writer) error {
	t, err := template.ParseFiles(path)
	if err != nil {
		log.Print(err)
		return err
	}

	err = t.Execute(w, params)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
