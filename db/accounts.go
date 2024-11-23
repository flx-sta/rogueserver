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
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pagefaultgames/rogueserver/defs"
)

func FetchAccounts(q string, limit int) ([]defs.Account, error) {
	log.Printf("Fetching accounts: %s, %d", q, limit)
	var accounts []defs.Account
	wildcardQuery := "%" + q + "%"
	dbQuery := `
	SELECT 
		a.username,
		a.registered,
		a.lastLoggedIn,
		a.lastActivity,
		a.banned,
		a.trainerId,
		a.secretId,
		a.discordId,
		a.googleId
	FROM accounts a
	WHERE a.username LIKE ?
	LIMIT ?`

	var rows *sql.Rows
	var err error

	rows, err = handle.Query(dbQuery, wildcardQuery, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item defs.Account

		err = rows.Scan(
			&item.Username,
			&item.Registered,
			&item.LastLoggedIn,
			&item.LastActivity,
			&item.Banned,
			&item.TrainerId,
			&item.SecretId,
			&item.DiscordId,
			&item.GoogleId,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, item)
	}

	return accounts, nil
}
