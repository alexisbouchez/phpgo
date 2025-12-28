package interpreter

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/alexisbouchez/phpgo/runtime"
)

// MySQLiObject represents a mysqli connection
type MySQLiObject struct {
	DB            *sql.DB
	Host          string
	Username      string
	Database      string
	Port          int
	AffectedRows  int64
	InsertID      int64
	Errno         int
	Error         string
	Connected     bool
	ServerInfo    string
	ServerVersion string
}

func NewMySQLi(host, username, password, database string, port int) *MySQLiObject {
	if port == 0 {
		port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", username, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return &MySQLiObject{
			Connected: false,
			Errno:     2002,
			Error:     err.Error(),
		}
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return &MySQLiObject{
			Connected: false,
			Errno:     2002,
			Error:     err.Error(),
		}
	}

	return &MySQLiObject{
		DB:            db,
		Host:          host,
		Username:      username,
		Database:      database,
		Port:          port,
		Connected:     true,
		Errno:         0,
		Error:         "",
		ServerInfo:    "MySQL",
		ServerVersion: "8.0",
	}
}

func (m *MySQLiObject) Type() string     { return "object" }
func (m *MySQLiObject) ToBool() bool     { return m.Connected }
func (m *MySQLiObject) ToInt() int64     { return 0 }
func (m *MySQLiObject) ToFloat() float64 { return 0 }
func (m *MySQLiObject) ToString() string { return "mysqli" }
func (m *MySQLiObject) Inspect() string {
	return fmt.Sprintf("object(mysqli)#%p", m)
}

func (m *MySQLiObject) Close() {
	if m.DB != nil {
		m.DB.Close()
		m.Connected = false
	}
}

func (m *MySQLiObject) Query(query string) runtime.Value {
	if !m.Connected || m.DB == nil {
		m.Errno = 2006
		m.Error = "MySQL server has gone away"
		return runtime.FALSE
	}

	// Check if it's a SELECT query
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "SHOW") ||
		strings.HasPrefix(trimmed, "DESCRIBE") ||
		strings.HasPrefix(trimmed, "EXPLAIN")

	if isSelect {
		rows, err := m.DB.Query(query)
		if err != nil {
			m.Errno = 1064
			m.Error = err.Error()
			return runtime.FALSE
		}
		return NewMySQLiResult(rows)
	}

	// For non-SELECT queries
	result, err := m.DB.Exec(query)
	if err != nil {
		m.Errno = 1064
		m.Error = err.Error()
		return runtime.FALSE
	}

	m.AffectedRows, _ = result.RowsAffected()
	m.InsertID, _ = result.LastInsertId()
	m.Errno = 0
	m.Error = ""

	return runtime.TRUE
}

func (m *MySQLiObject) Prepare(query string) runtime.Value {
	if !m.Connected || m.DB == nil {
		m.Errno = 2006
		m.Error = "MySQL server has gone away"
		return runtime.FALSE
	}

	stmt, err := m.DB.Prepare(query)
	if err != nil {
		m.Errno = 1064
		m.Error = err.Error()
		return runtime.FALSE
	}

	return NewMySQLiStmt(m, stmt, query)
}

func (m *MySQLiObject) RealEscapeString(str string) string {
	// Basic escaping for MySQL
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\x00", "\\0",
		"\n", "\\n",
		"\r", "\\r",
		"'", "\\'",
		"\"", "\\\"",
		"\x1a", "\\Z",
	)
	return replacer.Replace(str)
}

// MySQLiResultObject represents a mysqli query result
type MySQLiResultObject struct {
	Rows       *sql.Rows
	Columns    []string
	NumRows    int64
	CurrentRow int64
	cachedRows []map[string]interface{}
	cached     bool
}

func NewMySQLiResult(rows *sql.Rows) *MySQLiResultObject {
	cols, _ := rows.Columns()
	result := &MySQLiResultObject{
		Rows:       rows,
		Columns:    cols,
		CurrentRow: 0,
		cached:     false,
	}
	// Cache all rows to get num_rows
	result.cacheRows()
	return result
}

func (r *MySQLiResultObject) cacheRows() {
	if r.cached {
		return
	}

	r.cachedRows = make([]map[string]interface{}, 0)
	cols := r.Columns

	for r.Rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := r.Rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		r.cachedRows = append(r.cachedRows, row)
	}

	r.NumRows = int64(len(r.cachedRows))
	r.cached = true
}

func (r *MySQLiResultObject) Type() string     { return "object" }
func (r *MySQLiResultObject) ToBool() bool     { return true }
func (r *MySQLiResultObject) ToInt() int64     { return r.NumRows }
func (r *MySQLiResultObject) ToFloat() float64 { return float64(r.NumRows) }
func (r *MySQLiResultObject) ToString() string { return "mysqli_result" }
func (r *MySQLiResultObject) Inspect() string {
	return fmt.Sprintf("object(mysqli_result)#%p", r)
}

func (r *MySQLiResultObject) FetchAssoc() runtime.Value {
	if r.CurrentRow >= r.NumRows {
		return runtime.NULL
	}

	row := r.cachedRows[r.CurrentRow]
	r.CurrentRow++

	arr := runtime.NewArray()
	for col, val := range row {
		arr.Set(runtime.NewString(col), sqlValueToRuntime(val))
	}
	return arr
}

func (r *MySQLiResultObject) FetchRow() runtime.Value {
	if r.CurrentRow >= r.NumRows {
		return runtime.NULL
	}

	row := r.cachedRows[r.CurrentRow]
	r.CurrentRow++

	arr := runtime.NewArray()
	for i, col := range r.Columns {
		arr.Set(runtime.NewInt(int64(i)), sqlValueToRuntime(row[col]))
	}
	return arr
}

func (r *MySQLiResultObject) FetchArray(resultType int) runtime.Value {
	if r.CurrentRow >= r.NumRows {
		return runtime.NULL
	}

	row := r.cachedRows[r.CurrentRow]
	r.CurrentRow++

	arr := runtime.NewArray()

	// MYSQLI_BOTH = 3 (default), MYSQLI_ASSOC = 1, MYSQLI_NUM = 2
	if resultType == 1 || resultType == 3 {
		// Associative
		for col, val := range row {
			arr.Set(runtime.NewString(col), sqlValueToRuntime(val))
		}
	}
	if resultType == 2 || resultType == 3 {
		// Numeric
		for i, col := range r.Columns {
			arr.Set(runtime.NewInt(int64(i)), sqlValueToRuntime(row[col]))
		}
	}
	return arr
}

func (r *MySQLiResultObject) FetchAll(resultType int) runtime.Value {
	arr := runtime.NewArray()
	idx := int64(0)

	for r.CurrentRow < r.NumRows {
		row := r.FetchArray(resultType)
		arr.Set(runtime.NewInt(idx), row)
		idx++
	}

	return arr
}

func (r *MySQLiResultObject) FetchObject() runtime.Value {
	if r.CurrentRow >= r.NumRows {
		return runtime.NULL
	}

	row := r.cachedRows[r.CurrentRow]
	r.CurrentRow++

	// Create a simple stdClass-like object using array
	arr := runtime.NewArray()
	for col, val := range row {
		arr.Set(runtime.NewString(col), sqlValueToRuntime(val))
	}
	return arr
}

func (r *MySQLiResultObject) DataSeek(offset int64) bool {
	if offset < 0 || offset >= r.NumRows {
		return false
	}
	r.CurrentRow = offset
	return true
}

func (r *MySQLiResultObject) Free() {
	if r.Rows != nil {
		r.Rows.Close()
	}
}

// MySQLiStmtObject represents a mysqli prepared statement
type MySQLiStmtObject struct {
	Mysqli       *MySQLiObject
	Stmt         *sql.Stmt
	Query        string
	ParamCount   int
	BoundParams  []interface{}
	AffectedRows int64
	InsertID     int64
	Errno        int
	Error        string
}

func NewMySQLiStmt(mysqli *MySQLiObject, stmt *sql.Stmt, query string) *MySQLiStmtObject {
	// Count placeholders
	paramCount := strings.Count(query, "?")

	return &MySQLiStmtObject{
		Mysqli:      mysqli,
		Stmt:        stmt,
		Query:       query,
		ParamCount:  paramCount,
		BoundParams: make([]interface{}, paramCount),
	}
}

func (s *MySQLiStmtObject) Type() string     { return "object" }
func (s *MySQLiStmtObject) ToBool() bool     { return s.Stmt != nil }
func (s *MySQLiStmtObject) ToInt() int64     { return 0 }
func (s *MySQLiStmtObject) ToFloat() float64 { return 0 }
func (s *MySQLiStmtObject) ToString() string { return "mysqli_stmt" }
func (s *MySQLiStmtObject) Inspect() string {
	return fmt.Sprintf("object(mysqli_stmt)#%p", s)
}

func (s *MySQLiStmtObject) BindParam(types string, values []runtime.Value) bool {
	if len(values) != s.ParamCount {
		s.Errno = 2031
		s.Error = "No data supplied for parameters in prepared statement"
		return false
	}

	for i, val := range values {
		var typeChar byte = 's'
		if i < len(types) {
			typeChar = types[i]
		}

		switch typeChar {
		case 'i':
			s.BoundParams[i] = val.ToInt()
		case 'd':
			s.BoundParams[i] = val.ToFloat()
		case 'b':
			s.BoundParams[i] = []byte(val.ToString())
		default: // 's' and others
			s.BoundParams[i] = val.ToString()
		}
	}
	return true
}

func (s *MySQLiStmtObject) Execute() runtime.Value {
	if s.Stmt == nil {
		s.Errno = 2030
		s.Error = "No statement"
		return runtime.FALSE
	}

	// Check if it's a SELECT query
	trimmed := strings.TrimSpace(strings.ToUpper(s.Query))
	isSelect := strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "SHOW") ||
		strings.HasPrefix(trimmed, "DESCRIBE")

	if isSelect {
		rows, err := s.Stmt.Query(s.BoundParams...)
		if err != nil {
			s.Errno = 1064
			s.Error = err.Error()
			return runtime.FALSE
		}
		return NewMySQLiResult(rows)
	}

	result, err := s.Stmt.Exec(s.BoundParams...)
	if err != nil {
		s.Errno = 1064
		s.Error = err.Error()
		return runtime.FALSE
	}

	s.AffectedRows, _ = result.RowsAffected()
	s.InsertID, _ = result.LastInsertId()
	s.Mysqli.AffectedRows = s.AffectedRows
	s.Mysqli.InsertID = s.InsertID

	return runtime.TRUE
}

func (s *MySQLiStmtObject) Close() {
	if s.Stmt != nil {
		s.Stmt.Close()
		s.Stmt = nil
	}
}

// PDOObject represents a PDO connection
type PDOObject struct {
	DB           *sql.DB
	DSN          string
	DriverName   string
	InTransaction bool
	ErrorMode    int // 0=silent, 1=warning, 2=exception
	Errno        string
	Error        string
}

const (
	PDO_ERRMODE_SILENT    = 0
	PDO_ERRMODE_WARNING   = 1
	PDO_ERRMODE_EXCEPTION = 2
	PDO_FETCH_ASSOC       = 2
	PDO_FETCH_NUM         = 3
	PDO_FETCH_BOTH        = 4
	PDO_FETCH_OBJ         = 5
)

func NewPDO(dsn, username, password string) *PDOObject {
	// Parse DSN: mysql:host=localhost;dbname=test;port=3306
	parts := strings.SplitN(dsn, ":", 2)
	if len(parts) != 2 {
		return &PDOObject{
			Errno: "HY000",
			Error: "Invalid DSN",
		}
	}

	driver := parts[0]
	params := parts[1]

	if driver != "mysql" {
		return &PDOObject{
			Errno: "HY000",
			Error: fmt.Sprintf("could not find driver: %s", driver),
		}
	}

	// Parse parameters
	host := "localhost"
	port := "3306"
	dbname := ""

	for _, param := range strings.Split(params, ";") {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "host":
			host = value
		case "port":
			port = value
		case "dbname":
			dbname = value
		}
	}

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, dbname)

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return &PDOObject{
			Errno: "HY000",
			Error: err.Error(),
		}
	}

	if err := db.Ping(); err != nil {
		return &PDOObject{
			Errno: "HY000",
			Error: err.Error(),
		}
	}

	return &PDOObject{
		DB:         db,
		DSN:        dsn,
		DriverName: driver,
		ErrorMode:  PDO_ERRMODE_EXCEPTION,
	}
}

func (p *PDOObject) Type() string     { return "object" }
func (p *PDOObject) ToBool() bool     { return p.DB != nil }
func (p *PDOObject) ToInt() int64     { return 0 }
func (p *PDOObject) ToFloat() float64 { return 0 }
func (p *PDOObject) ToString() string { return "PDO" }
func (p *PDOObject) Inspect() string {
	return fmt.Sprintf("object(PDO)#%p", p)
}

func (p *PDOObject) Query(query string) runtime.Value {
	if p.DB == nil {
		return runtime.FALSE
	}

	rows, err := p.DB.Query(query)
	if err != nil {
		p.Errno = "HY000"
		p.Error = err.Error()
		if p.ErrorMode == PDO_ERRMODE_EXCEPTION {
			return runtime.NewError(fmt.Sprintf("SQLSTATE[%s]: %s", p.Errno, p.Error))
		}
		return runtime.FALSE
	}

	return NewPDOStatement(p, nil, rows)
}

func (p *PDOObject) Exec(query string) runtime.Value {
	if p.DB == nil {
		return runtime.FALSE
	}

	result, err := p.DB.Exec(query)
	if err != nil {
		p.Errno = "HY000"
		p.Error = err.Error()
		if p.ErrorMode == PDO_ERRMODE_EXCEPTION {
			return runtime.NewError(fmt.Sprintf("SQLSTATE[%s]: %s", p.Errno, p.Error))
		}
		return runtime.FALSE
	}

	affected, _ := result.RowsAffected()
	return runtime.NewInt(affected)
}

func (p *PDOObject) Prepare(query string) runtime.Value {
	if p.DB == nil {
		return runtime.FALSE
	}

	// Convert named placeholders to positional
	convertedQuery, paramOrder := convertNamedPlaceholders(query)

	stmt, err := p.DB.Prepare(convertedQuery)
	if err != nil {
		p.Errno = "HY000"
		p.Error = err.Error()
		if p.ErrorMode == PDO_ERRMODE_EXCEPTION {
			return runtime.NewError(fmt.Sprintf("SQLSTATE[%s]: %s", p.Errno, p.Error))
		}
		return runtime.FALSE
	}

	return NewPDOStatementPrepared(p, stmt, query, paramOrder)
}

func (p *PDOObject) BeginTransaction() bool {
	if p.DB == nil || p.InTransaction {
		return false
	}
	_, err := p.DB.Exec("BEGIN")
	if err != nil {
		return false
	}
	p.InTransaction = true
	return true
}

func (p *PDOObject) Commit() bool {
	if p.DB == nil || !p.InTransaction {
		return false
	}
	_, err := p.DB.Exec("COMMIT")
	if err != nil {
		return false
	}
	p.InTransaction = false
	return true
}

func (p *PDOObject) RollBack() bool {
	if p.DB == nil || !p.InTransaction {
		return false
	}
	_, err := p.DB.Exec("ROLLBACK")
	if err != nil {
		return false
	}
	p.InTransaction = false
	return true
}

func (p *PDOObject) LastInsertId() string {
	if p.DB == nil {
		return ""
	}
	var id int64
	p.DB.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return fmt.Sprintf("%d", id)
}

func (p *PDOObject) Quote(str string) string {
	// MySQL-style quoting
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\x00", "\\0",
		"\n", "\\n",
		"\r", "\\r",
		"'", "\\'",
		"\"", "\\\"",
		"\x1a", "\\Z",
	)
	return "'" + replacer.Replace(str) + "'"
}

// PDOStatementObject represents a PDO statement
type PDOStatementObject struct {
	PDO          *PDOObject
	Stmt         *sql.Stmt
	Rows         *sql.Rows
	Query        string
	Columns      []string
	BoundParams  map[string]interface{}
	ParamOrder   []string
	FetchMode    int
	cachedRows   []map[string]interface{}
	cached       bool
	currentRow   int64
	numRows      int64
}

func NewPDOStatement(pdo *PDOObject, stmt *sql.Stmt, rows *sql.Rows) *PDOStatementObject {
	cols, _ := rows.Columns()
	ps := &PDOStatementObject{
		PDO:         pdo,
		Stmt:        stmt,
		Rows:        rows,
		Columns:     cols,
		BoundParams: make(map[string]interface{}),
		FetchMode:   PDO_FETCH_BOTH,
	}
	ps.cacheRows()
	return ps
}

func NewPDOStatementPrepared(pdo *PDOObject, stmt *sql.Stmt, query string, paramOrder []string) *PDOStatementObject {
	return &PDOStatementObject{
		PDO:         pdo,
		Stmt:        stmt,
		Query:       query,
		BoundParams: make(map[string]interface{}),
		ParamOrder:  paramOrder,
		FetchMode:   PDO_FETCH_BOTH,
	}
}

func (s *PDOStatementObject) cacheRows() {
	if s.cached || s.Rows == nil {
		return
	}

	s.cachedRows = make([]map[string]interface{}, 0)

	for s.Rows.Next() {
		values := make([]interface{}, len(s.Columns))
		valuePtrs := make([]interface{}, len(s.Columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := s.Rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range s.Columns {
			row[col] = values[i]
		}
		s.cachedRows = append(s.cachedRows, row)
	}

	s.numRows = int64(len(s.cachedRows))
	s.cached = true
}

func (s *PDOStatementObject) Type() string     { return "object" }
func (s *PDOStatementObject) ToBool() bool     { return true }
func (s *PDOStatementObject) ToInt() int64     { return 0 }
func (s *PDOStatementObject) ToFloat() float64 { return 0 }
func (s *PDOStatementObject) ToString() string { return "PDOStatement" }
func (s *PDOStatementObject) Inspect() string {
	return fmt.Sprintf("object(PDOStatement)#%p", s)
}

func (s *PDOStatementObject) BindParam(param string, value runtime.Value) bool {
	s.BoundParams[param] = runtimeToSqlValue(value)
	return true
}

func (s *PDOStatementObject) BindValue(param string, value runtime.Value) bool {
	s.BoundParams[param] = runtimeToSqlValue(value)
	return true
}

func (s *PDOStatementObject) Execute(params []runtime.Value) bool {
	if s.Stmt == nil {
		return false
	}

	// Build parameter list in order
	var args []interface{}

	if len(params) > 0 {
		// Positional parameters passed directly
		for _, p := range params {
			args = append(args, runtimeToSqlValue(p))
		}
	} else if len(s.ParamOrder) > 0 {
		// Named parameters from bindParam
		for _, name := range s.ParamOrder {
			if val, ok := s.BoundParams[name]; ok {
				args = append(args, val)
			} else if val, ok := s.BoundParams[":"+name]; ok {
				args = append(args, val)
			} else {
				args = append(args, nil)
			}
		}
	} else {
		// Just bound params in order
		for i := 0; i < len(s.BoundParams); i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := s.BoundParams[key]; ok {
				args = append(args, val)
			}
		}
	}

	// Check if it's a SELECT query
	trimmed := strings.TrimSpace(strings.ToUpper(s.Query))
	isSelect := strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "SHOW") ||
		strings.HasPrefix(trimmed, "DESCRIBE")

	if isSelect {
		rows, err := s.Stmt.Query(args...)
		if err != nil {
			s.PDO.Errno = "HY000"
			s.PDO.Error = err.Error()
			return false
		}
		s.Rows = rows
		s.Columns, _ = rows.Columns()
		s.cached = false
		s.currentRow = 0
		s.cacheRows()
		return true
	}

	_, err := s.Stmt.Exec(args...)
	if err != nil {
		s.PDO.Errno = "HY000"
		s.PDO.Error = err.Error()
		return false
	}

	return true
}

func (s *PDOStatementObject) Fetch(fetchMode int) runtime.Value {
	if !s.cached {
		s.cacheRows()
	}

	if s.currentRow >= s.numRows {
		return runtime.FALSE
	}

	row := s.cachedRows[s.currentRow]
	s.currentRow++

	return s.formatRow(row, fetchMode)
}

func (s *PDOStatementObject) FetchAll(fetchMode int) runtime.Value {
	if !s.cached {
		s.cacheRows()
	}

	arr := runtime.NewArray()
	idx := int64(0)

	for s.currentRow < s.numRows {
		row := s.cachedRows[s.currentRow]
		s.currentRow++
		arr.Set(runtime.NewInt(idx), s.formatRow(row, fetchMode))
		idx++
	}

	return arr
}

func (s *PDOStatementObject) FetchColumn(columnIndex int) runtime.Value {
	if !s.cached {
		s.cacheRows()
	}

	if s.currentRow >= s.numRows {
		return runtime.FALSE
	}

	row := s.cachedRows[s.currentRow]
	s.currentRow++

	if columnIndex >= 0 && columnIndex < len(s.Columns) {
		col := s.Columns[columnIndex]
		return sqlValueToRuntime(row[col])
	}

	return runtime.FALSE
}

func (s *PDOStatementObject) RowCount() int64 {
	return s.numRows
}

func (s *PDOStatementObject) ColumnCount() int {
	return len(s.Columns)
}

func (s *PDOStatementObject) formatRow(row map[string]interface{}, fetchMode int) runtime.Value {
	if fetchMode == 0 {
		fetchMode = s.FetchMode
	}

	arr := runtime.NewArray()

	switch fetchMode {
	case PDO_FETCH_ASSOC:
		for col, val := range row {
			arr.Set(runtime.NewString(col), sqlValueToRuntime(val))
		}
	case PDO_FETCH_NUM:
		for i, col := range s.Columns {
			arr.Set(runtime.NewInt(int64(i)), sqlValueToRuntime(row[col]))
		}
	case PDO_FETCH_BOTH:
		for i, col := range s.Columns {
			val := sqlValueToRuntime(row[col])
			arr.Set(runtime.NewString(col), val)
			arr.Set(runtime.NewInt(int64(i)), val)
		}
	case PDO_FETCH_OBJ:
		for col, val := range row {
			arr.Set(runtime.NewString(col), sqlValueToRuntime(val))
		}
	default:
		// Default to BOTH
		for i, col := range s.Columns {
			val := sqlValueToRuntime(row[col])
			arr.Set(runtime.NewString(col), val)
			arr.Set(runtime.NewInt(int64(i)), val)
		}
	}

	return arr
}

func (s *PDOStatementObject) SetFetchMode(mode int) bool {
	s.FetchMode = mode
	return true
}

func (s *PDOStatementObject) CloseCursor() bool {
	if s.Rows != nil {
		s.Rows.Close()
		s.Rows = nil
	}
	s.cached = false
	s.cachedRows = nil
	s.currentRow = 0
	s.numRows = 0
	return true
}

// Helper functions

func sqlValueToRuntime(val interface{}) runtime.Value {
	if val == nil {
		return runtime.NULL
	}

	switch v := val.(type) {
	case int64:
		return runtime.NewInt(v)
	case int32:
		return runtime.NewInt(int64(v))
	case int:
		return runtime.NewInt(int64(v))
	case float64:
		return runtime.NewFloat(v)
	case float32:
		return runtime.NewFloat(float64(v))
	case bool:
		return runtime.NewBool(v)
	case string:
		return runtime.NewString(v)
	case []byte:
		return runtime.NewString(string(v))
	default:
		return runtime.NewString(fmt.Sprintf("%v", v))
	}
}

func runtimeToSqlValue(val runtime.Value) interface{} {
	switch v := val.(type) {
	case *runtime.Int:
		return v.Value
	case *runtime.Float:
		return v.Value
	case *runtime.Bool:
		if v.Value {
			return 1
		}
		return 0
	case *runtime.String:
		return v.Value
	case *runtime.Null:
		return nil
	default:
		return val.ToString()
	}
}

func convertNamedPlaceholders(query string) (string, []string) {
	// Convert :name placeholders to ?
	var paramOrder []string
	result := query

	// Find all :name patterns
	i := 0
	for i < len(result) {
		if result[i] == ':' && i+1 < len(result) && isAlphaNumeric(result[i+1]) {
			// Find end of parameter name
			j := i + 1
			for j < len(result) && isAlphaNumeric(result[j]) {
				j++
			}
			paramName := result[i+1 : j]
			paramOrder = append(paramOrder, paramName)
			result = result[:i] + "?" + result[j:]
		}
		i++
	}

	return result, paramOrder
}

func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// Check if class name is a database class
func isDatabaseClass(name string) bool {
	switch name {
	case "mysqli", "mysqli_result", "mysqli_stmt", "PDO", "PDOStatement":
		return true
	}
	return false
}

// handleDatabaseNew creates a new database object
func (i *Interpreter) handleDatabaseNew(className string, args []runtime.Value) runtime.Value {
	switch className {
	case "mysqli":
		host := "localhost"
		username := ""
		password := ""
		database := ""
		port := 3306

		if len(args) >= 1 && args[0] != runtime.NULL {
			host = args[0].ToString()
		}
		if len(args) >= 2 && args[1] != runtime.NULL {
			username = args[1].ToString()
		}
		if len(args) >= 3 && args[2] != runtime.NULL {
			password = args[2].ToString()
		}
		if len(args) >= 4 && args[3] != runtime.NULL {
			database = args[3].ToString()
		}
		if len(args) >= 5 && args[4] != runtime.NULL {
			port = int(args[4].ToInt())
		}

		return NewMySQLi(host, username, password, database, port)

	case "PDO":
		if len(args) < 1 {
			return runtime.NewError("PDO::__construct() expects at least 1 parameter")
		}
		dsn := args[0].ToString()
		username := ""
		password := ""
		if len(args) >= 2 && args[1] != runtime.NULL {
			username = args[1].ToString()
		}
		if len(args) >= 3 && args[2] != runtime.NULL {
			password = args[2].ToString()
		}
		// Options (arg 4) are ignored for now

		pdo := NewPDO(dsn, username, password)
		if pdo.Error != "" {
			return runtime.NewError(fmt.Sprintf("PDO::__construct(): %s", pdo.Error))
		}
		return pdo
	}

	return runtime.NewError(fmt.Sprintf("unknown database class: %s", className))
}

// callDatabaseMethod handles method calls on database objects
func (i *Interpreter) callDatabaseMethod(obj runtime.Value, methodName string, args []runtime.Value) runtime.Value {
	switch o := obj.(type) {
	case *MySQLiObject:
		return i.callMySQLiMethod(o, methodName, args)
	case *MySQLiResultObject:
		return i.callMySQLiResultMethod(o, methodName, args)
	case *MySQLiStmtObject:
		return i.callMySQLiStmtMethod(o, methodName, args)
	case *PDOObject:
		return i.callPDOMethod(o, methodName, args)
	case *PDOStatementObject:
		return i.callPDOStatementMethod(o, methodName, args)
	}
	return runtime.NewError("unknown database object type")
}

func (i *Interpreter) callMySQLiMethod(m *MySQLiObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "query":
		if len(args) < 1 {
			return runtime.NewError("mysqli::query() expects exactly 1 parameter")
		}
		return m.Query(args[0].ToString())

	case "prepare":
		if len(args) < 1 {
			return runtime.NewError("mysqli::prepare() expects exactly 1 parameter")
		}
		return m.Prepare(args[0].ToString())

	case "real_escape_string", "escape_string":
		if len(args) < 1 {
			return runtime.NewError("mysqli::real_escape_string() expects exactly 1 parameter")
		}
		return runtime.NewString(m.RealEscapeString(args[0].ToString()))

	case "close":
		m.Close()
		return runtime.TRUE

	case "ping":
		if m.DB != nil && m.DB.Ping() == nil {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "begin_transaction":
		if m.DB == nil {
			return runtime.FALSE
		}
		_, err := m.DB.Exec("BEGIN")
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE

	case "commit":
		if m.DB == nil {
			return runtime.FALSE
		}
		_, err := m.DB.Exec("COMMIT")
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE

	case "rollback":
		if m.DB == nil {
			return runtime.FALSE
		}
		_, err := m.DB.Exec("ROLLBACK")
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE

	case "autocommit":
		if m.DB == nil || len(args) < 1 {
			return runtime.FALSE
		}
		mode := "ON"
		if !args[0].ToBool() {
			mode = "OFF"
		}
		_, err := m.DB.Exec("SET autocommit = " + mode)
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE

	case "select_db":
		if m.DB == nil || len(args) < 1 {
			return runtime.FALSE
		}
		_, err := m.DB.Exec("USE " + args[0].ToString())
		if err != nil {
			return runtime.FALSE
		}
		m.Database = args[0].ToString()
		return runtime.TRUE
	}

	return runtime.NewError(fmt.Sprintf("undefined method: mysqli::%s", methodName))
}

func (i *Interpreter) callMySQLiResultMethod(r *MySQLiResultObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "fetch_assoc":
		return r.FetchAssoc()

	case "fetch_row":
		return r.FetchRow()

	case "fetch_array":
		resultType := 3 // MYSQLI_BOTH
		if len(args) >= 1 {
			resultType = int(args[0].ToInt())
		}
		return r.FetchArray(resultType)

	case "fetch_all":
		resultType := 1 // MYSQLI_ASSOC
		if len(args) >= 1 {
			resultType = int(args[0].ToInt())
		}
		return r.FetchAll(resultType)

	case "fetch_object":
		return r.FetchObject()

	case "data_seek":
		if len(args) < 1 {
			return runtime.FALSE
		}
		if r.DataSeek(args[0].ToInt()) {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "free", "free_result", "close":
		r.Free()
		return runtime.NULL
	}

	return runtime.NewError(fmt.Sprintf("undefined method: mysqli_result::%s", methodName))
}

func (i *Interpreter) callMySQLiStmtMethod(s *MySQLiStmtObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "bind_param":
		if len(args) < 2 {
			return runtime.FALSE
		}
		types := args[0].ToString()
		values := args[1:]
		if s.BindParam(types, values) {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "execute":
		return s.Execute()

	case "get_result":
		// For SELECT queries, execute returns the result directly
		// This method is for compatibility
		return s.Execute()

	case "close":
		s.Close()
		return runtime.TRUE

	case "fetch":
		// Not typically used, but return false
		return runtime.FALSE
	}

	return runtime.NewError(fmt.Sprintf("undefined method: mysqli_stmt::%s", methodName))
}

func (i *Interpreter) callPDOMethod(p *PDOObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "query":
		if len(args) < 1 {
			return runtime.FALSE
		}
		return p.Query(args[0].ToString())

	case "exec":
		if len(args) < 1 {
			return runtime.FALSE
		}
		return p.Exec(args[0].ToString())

	case "prepare":
		if len(args) < 1 {
			return runtime.FALSE
		}
		return p.Prepare(args[0].ToString())

	case "beginTransaction":
		if p.BeginTransaction() {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "commit":
		if p.Commit() {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "rollBack":
		if p.RollBack() {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "inTransaction":
		return runtime.NewBool(p.InTransaction)

	case "lastInsertId":
		return runtime.NewString(p.LastInsertId())

	case "quote":
		if len(args) < 1 {
			return runtime.FALSE
		}
		return runtime.NewString(p.Quote(args[0].ToString()))

	case "errorCode":
		return runtime.NewString(p.Errno)

	case "errorInfo":
		arr := runtime.NewArray()
		arr.Set(runtime.NewInt(0), runtime.NewString(p.Errno))
		arr.Set(runtime.NewInt(1), runtime.NewString(""))
		arr.Set(runtime.NewInt(2), runtime.NewString(p.Error))
		return arr

	case "setAttribute":
		// Handle common attributes
		if len(args) >= 2 {
			attr := int(args[0].ToInt())
			if attr == 3 { // PDO::ATTR_ERRMODE
				p.ErrorMode = int(args[1].ToInt())
			}
		}
		return runtime.TRUE

	case "getAttribute":
		if len(args) < 1 {
			return runtime.NULL
		}
		attr := int(args[0].ToInt())
		switch attr {
		case 3: // PDO::ATTR_ERRMODE
			return runtime.NewInt(int64(p.ErrorMode))
		case 4: // PDO::ATTR_DRIVER_NAME
			return runtime.NewString(p.DriverName)
		}
		return runtime.NULL
	}

	return runtime.NewError(fmt.Sprintf("undefined method: PDO::%s", methodName))
}

func (i *Interpreter) callPDOStatementMethod(s *PDOStatementObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "execute":
		if s.Execute(args) {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "bindParam", "bindValue":
		if len(args) < 2 {
			return runtime.FALSE
		}
		param := args[0].ToString()
		if s.BindParam(param, args[1]) {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "fetch":
		fetchMode := 0
		if len(args) >= 1 {
			fetchMode = int(args[0].ToInt())
		}
		return s.Fetch(fetchMode)

	case "fetchAll":
		fetchMode := 0
		if len(args) >= 1 {
			fetchMode = int(args[0].ToInt())
		}
		return s.FetchAll(fetchMode)

	case "fetchColumn":
		columnIndex := 0
		if len(args) >= 1 {
			columnIndex = int(args[0].ToInt())
		}
		return s.FetchColumn(columnIndex)

	case "rowCount":
		return runtime.NewInt(s.RowCount())

	case "columnCount":
		return runtime.NewInt(int64(s.ColumnCount()))

	case "setFetchMode":
		if len(args) < 1 {
			return runtime.FALSE
		}
		if s.SetFetchMode(int(args[0].ToInt())) {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "closeCursor":
		if s.CloseCursor() {
			return runtime.TRUE
		}
		return runtime.FALSE

	case "errorCode":
		return runtime.NewString(s.PDO.Errno)

	case "errorInfo":
		arr := runtime.NewArray()
		arr.Set(runtime.NewInt(0), runtime.NewString(s.PDO.Errno))
		arr.Set(runtime.NewInt(1), runtime.NewString(""))
		arr.Set(runtime.NewInt(2), runtime.NewString(s.PDO.Error))
		return arr
	}

	return runtime.NewError(fmt.Sprintf("undefined method: PDOStatement::%s", methodName))
}

// Property access for database objects
func (i *Interpreter) getDatabaseProperty(obj runtime.Value, prop string) runtime.Value {
	switch o := obj.(type) {
	case *MySQLiObject:
		switch prop {
		case "affected_rows":
			return runtime.NewInt(o.AffectedRows)
		case "insert_id":
			return runtime.NewInt(o.InsertID)
		case "errno":
			return runtime.NewInt(int64(o.Errno))
		case "error":
			return runtime.NewString(o.Error)
		case "connect_errno":
			if !o.Connected {
				return runtime.NewInt(int64(o.Errno))
			}
			return runtime.NewInt(0)
		case "connect_error":
			if !o.Connected {
				return runtime.NewString(o.Error)
			}
			return runtime.NULL
		case "host_info":
			return runtime.NewString(fmt.Sprintf("%s via TCP/IP", o.Host))
		case "server_info":
			return runtime.NewString(o.ServerInfo)
		case "server_version":
			return runtime.NewString(o.ServerVersion)
		}
	case *MySQLiResultObject:
		switch prop {
		case "num_rows":
			return runtime.NewInt(o.NumRows)
		case "current_field":
			return runtime.NewInt(0)
		case "field_count":
			return runtime.NewInt(int64(len(o.Columns)))
		}
	case *MySQLiStmtObject:
		switch prop {
		case "affected_rows":
			return runtime.NewInt(o.AffectedRows)
		case "insert_id":
			return runtime.NewInt(o.InsertID)
		case "errno":
			return runtime.NewInt(int64(o.Errno))
		case "error":
			return runtime.NewString(o.Error)
		case "param_count":
			return runtime.NewInt(int64(o.ParamCount))
		case "num_rows":
			return runtime.NewInt(0) // Would need result to know
		}
	}
	return runtime.NULL
}

// ============================================================================
// MySQLi procedural interface functions
// ============================================================================

func (i *Interpreter) builtinMysqliConnect(args ...runtime.Value) runtime.Value {
	host := "localhost"
	username := ""
	password := ""
	database := ""
	port := 3306

	if len(args) >= 1 && args[0] != runtime.NULL {
		host = args[0].ToString()
	}
	if len(args) >= 2 && args[1] != runtime.NULL {
		username = args[1].ToString()
	}
	if len(args) >= 3 && args[2] != runtime.NULL {
		password = args[2].ToString()
	}
	if len(args) >= 4 && args[3] != runtime.NULL {
		database = args[3].ToString()
	}
	if len(args) >= 5 && args[4] != runtime.NULL {
		port = int(args[4].ToInt())
	}

	mysqli := NewMySQLi(host, username, password, database, port)
	if !mysqli.Connected {
		return runtime.FALSE
	}
	return mysqli
}

func (i *Interpreter) builtinMysqliClose(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	if mysqli, ok := args[0].(*MySQLiObject); ok {
		mysqli.Close()
		return runtime.TRUE
	}
	return runtime.FALSE
}

func (i *Interpreter) builtinMysqliQuery(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	return mysqli.Query(args[1].ToString())
}

func (i *Interpreter) builtinMysqliPrepare(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	return mysqli.Prepare(args[1].ToString())
}

func (i *Interpreter) builtinMysqliStmtBindParam(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.FALSE
	}
	stmt, ok := args[0].(*MySQLiStmtObject)
	if !ok {
		return runtime.FALSE
	}
	types := args[1].ToString()
	values := args[2:]
	if stmt.BindParam(types, values) {
		return runtime.TRUE
	}
	return runtime.FALSE
}

func (i *Interpreter) builtinMysqliStmtExecute(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	stmt, ok := args[0].(*MySQLiStmtObject)
	if !ok {
		return runtime.FALSE
	}
	result := stmt.Execute()
	if _, isError := result.(*runtime.Error); isError {
		return runtime.FALSE
	}
	if result == runtime.FALSE {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliStmtGetResult(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	stmt, ok := args[0].(*MySQLiStmtObject)
	if !ok {
		return runtime.FALSE
	}
	return stmt.Execute()
}

func (i *Interpreter) builtinMysqliStmtClose(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	stmt, ok := args[0].(*MySQLiStmtObject)
	if !ok {
		return runtime.FALSE
	}
	stmt.Close()
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliFetchAssoc(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NULL
	}
	return result.FetchAssoc()
}

func (i *Interpreter) builtinMysqliFetchRow(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NULL
	}
	return result.FetchRow()
}

func (i *Interpreter) builtinMysqliFetchArray(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NULL
	}
	resultType := 3 // MYSQLI_BOTH
	if len(args) >= 2 {
		resultType = int(args[1].ToInt())
	}
	return result.FetchArray(resultType)
}

func (i *Interpreter) builtinMysqliFetchAll(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NewArray()
	}
	resultType := 1 // MYSQLI_ASSOC
	if len(args) >= 2 {
		resultType = int(args[1].ToInt())
	}
	return result.FetchAll(resultType)
}

func (i *Interpreter) builtinMysqliFetchObject(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NULL
	}
	return result.FetchObject()
}

func (i *Interpreter) builtinMysqliNumRows(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.FALSE
	}
	return runtime.NewInt(result.NumRows)
}

func (i *Interpreter) builtinMysqliAffectedRows(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(-1)
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.NewInt(-1)
	}
	return runtime.NewInt(mysqli.AffectedRows)
}

func (i *Interpreter) builtinMysqliInsertId(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(mysqli.InsertID)
}

func (i *Interpreter) builtinMysqliRealEscapeString(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	return runtime.NewString(mysqli.RealEscapeString(args[1].ToString()))
}

func (i *Interpreter) builtinMysqliError(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.NewString("")
	}
	return runtime.NewString(mysqli.Error)
}

func (i *Interpreter) builtinMysqliErrno(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(int64(mysqli.Errno))
}

func (i *Interpreter) builtinMysqliFreeResult(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.NULL
	}
	result.Free()
	return runtime.NULL
}

func (i *Interpreter) builtinMysqliDataSeek(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	result, ok := args[0].(*MySQLiResultObject)
	if !ok {
		return runtime.FALSE
	}
	if result.DataSeek(args[1].ToInt()) {
		return runtime.TRUE
	}
	return runtime.FALSE
}

func (i *Interpreter) builtinMysqliSelectDb(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB == nil {
		return runtime.FALSE
	}
	_, err := mysqli.DB.Exec("USE " + args[1].ToString())
	if err != nil {
		return runtime.FALSE
	}
	mysqli.Database = args[1].ToString()
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliPing(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB != nil && mysqli.DB.Ping() == nil {
		return runtime.TRUE
	}
	return runtime.FALSE
}

func (i *Interpreter) builtinMysqliBeginTransaction(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB == nil {
		return runtime.FALSE
	}
	_, err := mysqli.DB.Exec("BEGIN")
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliCommit(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB == nil {
		return runtime.FALSE
	}
	_, err := mysqli.DB.Exec("COMMIT")
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliRollback(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB == nil {
		return runtime.FALSE
	}
	_, err := mysqli.DB.Exec("ROLLBACK")
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func (i *Interpreter) builtinMysqliAutocommit(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	mysqli, ok := args[0].(*MySQLiObject)
	if !ok {
		return runtime.FALSE
	}
	if mysqli.DB == nil {
		return runtime.FALSE
	}
	mode := "ON"
	if !args[1].ToBool() {
		mode = "OFF"
	}
	_, err := mysqli.DB.Exec("SET autocommit = " + mode)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

// ============================================================================
// Database constants registration
// ============================================================================

func (i *Interpreter) registerDatabaseConstants() {
	// MySQLi constants
	i.env.DefineConstant("MYSQLI_ASSOC", runtime.NewInt(1))
	i.env.DefineConstant("MYSQLI_NUM", runtime.NewInt(2))
	i.env.DefineConstant("MYSQLI_BOTH", runtime.NewInt(3))

	// PDO constants
	i.env.DefineConstant("PDO::ATTR_ERRMODE", runtime.NewInt(3))
	i.env.DefineConstant("PDO::ERRMODE_SILENT", runtime.NewInt(0))
	i.env.DefineConstant("PDO::ERRMODE_WARNING", runtime.NewInt(1))
	i.env.DefineConstant("PDO::ERRMODE_EXCEPTION", runtime.NewInt(2))
	i.env.DefineConstant("PDO::ATTR_DEFAULT_FETCH_MODE", runtime.NewInt(19))
	i.env.DefineConstant("PDO::ATTR_DRIVER_NAME", runtime.NewInt(16))
	i.env.DefineConstant("PDO::FETCH_ASSOC", runtime.NewInt(2))
	i.env.DefineConstant("PDO::FETCH_NUM", runtime.NewInt(3))
	i.env.DefineConstant("PDO::FETCH_BOTH", runtime.NewInt(4))
	i.env.DefineConstant("PDO::FETCH_OBJ", runtime.NewInt(5))
	i.env.DefineConstant("PDO::FETCH_COLUMN", runtime.NewInt(7))
	i.env.DefineConstant("PDO::PARAM_NULL", runtime.NewInt(0))
	i.env.DefineConstant("PDO::PARAM_INT", runtime.NewInt(1))
	i.env.DefineConstant("PDO::PARAM_STR", runtime.NewInt(2))
	i.env.DefineConstant("PDO::PARAM_BOOL", runtime.NewInt(5))
}
