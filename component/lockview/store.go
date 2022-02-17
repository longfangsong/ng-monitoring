package lockview

import (
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
)

type Store struct {
	documentDB *genji.DB
}

func NewStore(documentDB *genji.DB) Store {
	result := Store{documentDB}
	result.init()
	return result
}

func (s *Store) init() {
	createTableStmts := []string{
		"CREATE TABLE IF NOT EXISTS TIDB_TRX (id INTEGER PRIMARY KEY, ALL_SQL_DIGESTS TEXT, EXEC_TIME INTEGER)",
	}
	for _, stmt := range createTableStmts {
		if err := s.documentDB.Exec(stmt); err != nil {
			panic(err)
		}
	}
}

func (s *Store) Insert(id uint64, allSqlDigest string, execTime uint64) error {
	prepareStmt := "INSERT INTO TIDB_TRX(id, ALL_SQL_DIGESTS, EXEC_TIME) VALUES (?, ?, ?) ON CONFLICT DO NOTHING"
	prepare, err := s.documentDB.Prepare(prepareStmt)
	if err != nil {
		panic(err)
	}

	return prepare.Exec(id, allSqlDigest, execTime)
}

func (s *Store) Select(idStart uint64, idEnd uint64) []Model {
	prepareStmt := "SELECT id, ALL_SQL_DIGESTS, EXEC_TIME from TIDB_TRX WHERE id >= ? AND id <= ?"
	prepare, err := s.documentDB.Prepare(prepareStmt)
	if err != nil {
		panic(err)
	}
	queryResult, err := prepare.Query(idStart, idEnd)
	if err != nil {
		panic(err)
	}
	result := []Model{}
	err = queryResult.Iterate(func(d types.Document) error {
		m := Model{}
		err := document.Scan(d, &m.Id, &m.AllSqlDigest, &m.ExecTime)
		if err == nil {
			result = append(result, m)
		}
		return err
	})
	if err != nil {
		panic(err)
	}
	return result
}
