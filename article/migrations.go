package article

import migrate "github.com/rubenv/sql-migrate"

func Migrations() []*migrate.Migration {
	return []*migrate.Migration{
		{
			Id: "0001_initial",
			Up: []string{
				`CREATE TABLE article (
					id          SERIAL                      NOT NULL,
					title       character varying(256)      NOT NULL,
					slug        character varying(128)      NOT NULL,
					PRIMARY KEY (id)
				);`,
			},
		},
	}
}
