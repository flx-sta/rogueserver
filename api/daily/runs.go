/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package daily

import (
	"log"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func DailyRuns(page int, limit int, searchQuery string) ([]defs.DailyRun, error) {
	if limit > 100 {
		log.Printf("Limit is greater than 100, setting limit to 100")
		limit = 100
	} else if limit < 10 {
		log.Printf("Limit is less than 10, setting limit to 10")
		limit = 10
	}

	log.Printf("Fetching daily runs: %d, %d, %s", page, limit, searchQuery)

	rankings, err := db.FetchDailyRuns(page, limit, searchQuery)
	if err != nil {
		return rankings, err
	}

	return rankings, nil
}

func DailyRunsTotalCount(includeDeleted bool) (int, error) {
	log.Print("Fetching daily runs total count")

	return db.FetchDailyRunsTotalCount(includeDeleted)
}

func SoftDeleteDailyRun(username string, date string, discordId string) (bool, error) {
	log.Printf("Soft deleting daily run: %s, %s, %s", username, date, discordId)

	return db.SoftDeleteDailyRun(username, date, discordId)
}

func RestoreDeletedDailyRun(username string, date string) (bool, error) {
	log.Printf("Restoring deleted daily run: %s, %s", username, date)

	return db.RestoreDeletedDailyRun(username, date)
}
