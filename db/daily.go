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

package db

import (
	"database/sql"
	"math"

	"github.com/pagefaultgames/rogueserver/defs"
)

func TryAddDailyRun(seed string) (string, error) {
	var actualSeed string
	err := handle.QueryRow("INSERT INTO dailyRuns (seed, date) VALUES (?, UTC_DATE()) ON DUPLICATE KEY UPDATE date = date RETURNING seed", seed).Scan(&actualSeed)
	if err != nil {
		return "", err
	}

	return actualSeed, nil
}

func GetDailyRunSeed() (string, error) {
	var seed string
	err := handle.QueryRow("SELECT seed FROM dailyRuns WHERE date = UTC_DATE()").Scan(&seed)
	if err != nil {
		return "", err
	}

	return seed, nil
}

func AddOrUpdateAccountDailyRun(uuid []byte, score int, wave int) error {
	_, err := handle.Exec("INSERT INTO accountDailyRuns (uuid, date, score, wave, timestamp) VALUES (?, UTC_DATE(), ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE score = GREATEST(score, ?), wave = GREATEST(wave, ?), timestamp = IF(score < ?, UTC_TIMESTAMP(), timestamp)", uuid, score, wave, score, wave, score)
	if err != nil {
		return err
	}

	return nil
}

func FetchRankings(category int, page int) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking

	offset := (page - 1) * 10

	var query string
	switch category {
	case 0:
		query = "SELECT a.username, adr.score, adr.wave FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE() AND a.banned = 0 AND adr.deleted = 0 LIMIT 10 OFFSET ?"
	case 1:
		query = "SELECT RANK() OVER (ORDER BY SUM(adr.score) DESC, adr.timestamp), a.username, SUM(adr.score), 0 FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY) AND a.banned = 0 AND adr.deleted = 0 GROUP BY a.username ORDER BY 1 LIMIT 10 OFFSET ?"
	}

	results, err := handle.Query(query, offset)
	if err != nil {
		return rankings, err
	}

	defer results.Close()

	for results.Next() {
		var ranking defs.DailyRanking
		err = results.Scan(&ranking.Rank, &ranking.Username, &ranking.Score, &ranking.Wave)
		if err != nil {
			return rankings, err
		}

		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

func FetchRankingPageCount(category int) (int, error) {
	var query string
	switch category {
	case 0:
		query = "SELECT COUNT(a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE() AND adr.deleted = 0"
	case 1:
		query = "SELECT COUNT(DISTINCT a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY) AND adr.deleted = 0"
	}

	var recordCount int
	err := handle.QueryRow(query).Scan(&recordCount)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(float64(recordCount) / 10)), nil
}

func FetchDailyRunsTotalCount(includeDeleted bool) (int, error) {
	var totalCount int
	dbQuery := `
		SELECT COUNT(*) 
		FROM accountDailyRuns
	`

	if !includeDeleted {
		dbQuery += " WHERE deleted = 0"
	}

	row := handle.QueryRow(dbQuery)
	err := row.Scan(&totalCount)
	if err != nil {
		return -1, err
	}

	return totalCount, nil
}

var dailyRunSelectQuery = `
	SELECT 
		a.username, 
		adr.date, 
		adr.score, 
		adr.wave, 
		adr.deleted, 
		adr.deletedAt, 
		adr.deletedByDiscordId
	FROM accountDailyRuns adr 
	JOIN accounts a ON adr.uuid = a.uuid
	`

var dailyRunUpdateSelectQuery = `
	UPDATE accountDailyRuns adr
	JOIN accounts a ON adr.uuid = a.uuid
`

func ScanDailyRunRows(rows *sql.Rows, dailyRun *defs.DailyRun) error {
	return rows.Scan(
		&dailyRun.Username,
		&dailyRun.Date,
		&dailyRun.Score,
		&dailyRun.Wave,
		&dailyRun.Deleted,
		&dailyRun.DeletedAt,
		&dailyRun.DeletedByDiscordId,
	)
}

func FetchDailyRuns(page int, limit int, searchQuery string) ([]defs.DailyRun, error) {
	var dailyRuns []defs.DailyRun
	offset := (page - 1) * 10
	wildcardQuery := "%" + searchQuery + "%"

	dbQuery := dailyRunSelectQuery

	if searchQuery != "" {
		dbQuery += `
		WHERE 
			a.username LIKE ? 
			OR adr.score LIKE ? 
			OR adr.wave LIKE ? 
			OR adr.date LIKE ? 
		`
	}

	// Add LIMIT and OFFSET (this is always included)
	dbQuery += `LIMIT ? OFFSET ?`

	var rows *sql.Rows
	var err error

	if searchQuery == "" {
		rows, err = handle.Query(dbQuery, limit, offset)
	} else {
		rows, err = handle.Query(dbQuery, wildcardQuery, wildcardQuery, wildcardQuery, wildcardQuery, limit, offset)
	}

	if err != nil {
		return dailyRuns, err
	}

	defer rows.Close()

	for rows.Next() {
		var item defs.DailyRun

		err = rows.Scan(
			&item.Username,
			&item.Date,
			&item.Score,
			&item.Wave,
			&item.Deleted,
			&item.DeletedAt,
			&item.DeletedByDiscordId,
		)
		if err != nil {
			return dailyRuns, err
		}

		dailyRuns = append(dailyRuns, item)
	}

	return dailyRuns, nil
}

func FetchDailyRun(username string, date string) (*defs.DailyRun, error) {
	dbQuery := dailyRunSelectQuery + `
							WHERE a.username = ? 
							AND adr.date = ?;`

	var dailyRun defs.DailyRun
	row := handle.QueryRow(dbQuery, username, date)

	err := row.Scan(
		&dailyRun.Username,
		&dailyRun.Date,
		&dailyRun.Score,
		&dailyRun.Wave,
		&dailyRun.Deleted,
		&dailyRun.DeletedAt,
		&dailyRun.DeletedByDiscordId,
	)
	if err != nil {
		return nil, err
	}

	return &dailyRun, nil
}

func SoftDeleteDailyRun(username string, date string, discordId string) (bool, error) {
	ranking, err := FetchDailyRun(username, date)
	if err != nil {
		return false, err
	}

	if (ranking == nil) || (ranking.Deleted == 1) {
		// already deleted
		return false, nil
	}

	dbQuery := dailyRunUpdateSelectQuery + `
		SET deleted = 1, 
				deletedAt = UTC_TIMESTAMP(), 
				deletedByDiscordId = ? 
		WHERE a.username = ?
			AND date = ?;
	`

	_, err = handle.Exec(dbQuery, discordId, username, date)
	if err != nil {
		return false, err
	}

	return true, nil
}

func RestoreDeletedDailyRun(username string, date string) (bool, error) {
	ranking, err := FetchDailyRun(username, date)
	if err != nil || ranking == nil {
		return false, err
	}

	if ranking.Deleted == 0 {
		// run isn't deleted
		return false, nil
	}

	dbQuery := dailyRunUpdateSelectQuery + `
		SET deleted = 0, 
				deletedAt = null, 
				deletedByDiscordId = null 
		WHERE a.username = ?
			AND date = ?;
	`

	_, err = handle.Exec(dbQuery, username, date)
	if err != nil {
		return false, err
	}

	return true, nil
}
