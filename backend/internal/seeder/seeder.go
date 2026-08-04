package seeder

import "github.com/jmoiron/sqlx"

// Seeder é uma função que executa um conjunto de inserts no banco.
type Seeder func(db *sqlx.DB) error

// Run executa os seeders na ordem informada, parando no primeiro erro.
func Run(db *sqlx.DB, seeders ...Seeder) error {
	for _, s := range seeders {
		if err := s(db); err != nil {
			return err
		}
	}
	return nil
}
