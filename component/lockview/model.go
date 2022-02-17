package lockview

type Model struct {
	Id           uint64 `json:"id"`
	AllSqlDigest string `json:"all_sql_digest"`
	ExecTime     uint64 `json:"exec_time"`
}
