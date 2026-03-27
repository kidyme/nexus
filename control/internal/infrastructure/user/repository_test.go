package user

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
	"testing"

	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

type stubExecResult struct {
	rowsAffected int64
	err          error
}

type stubState struct {
	execResults   []stubExecResult
	execArgCounts []int
	beginCalls    int
	commitCalls   int
	rollbackCalls int
}

type stubDriver struct{}
type stubConn struct {
	state *stubState
}
type stubTx struct {
	state *stubState
}
type stubResult struct {
	rowsAffected int64
}
type stubStmt struct{}

var (
	registerStubDriverOnce sync.Once
	stubStates             sync.Map
)

func openStubDB(t *testing.T, state *stubState) *sql.DB {
	t.Helper()
	registerStubDriverOnce.Do(func() {
		sql.Register("user-repository-stub", &stubDriver{})
	})
	dsn := fmt.Sprintf("%p", state)
	stubStates.Store(dsn, state)
	t.Cleanup(func() {
		stubStates.Delete(dsn)
	})
	db, err := sql.Open("user-repository-stub", dsn)
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func (d *stubDriver) Open(name string) (driver.Conn, error) {
	state, ok := stubStates.Load(name)
	if !ok {
		return nil, fmt.Errorf("stub state not found for %s", name)
	}
	return &stubConn{state: state.(*stubState)}, nil
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) { return stubStmt{}, nil }
func (c *stubConn) Close() error                        { return nil }
func (c *stubConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *stubConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.beginCalls++
	return &stubTx{state: c.state}, nil
}

func (c *stubConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execArgCounts = append(c.state.execArgCounts, len(args))
	if len(c.state.execResults) == 0 {
		return nil, fmt.Errorf("no stub exec result configured")
	}
	result := c.state.execResults[0]
	c.state.execResults = c.state.execResults[1:]
	if result.err != nil {
		return nil, result.err
	}
	return stubResult{rowsAffected: result.rowsAffected}, nil
}

func (t stubTx) Commit() error {
	t.state.commitCalls++
	return nil
}

func (t stubTx) Rollback() error {
	t.state.rollbackCalls++
	return nil
}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (r stubResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (stubStmt) Close() error  { return nil }
func (stubStmt) NumInput() int { return -1 }
func (stubStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("not implemented")
}
func (stubStmt) Query([]driver.Value) (driver.Rows, error) { return nil, fmt.Errorf("not implemented") }
func (stubStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	return nil, fmt.Errorf("not implemented")
}
func (stubStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return nil, fmt.Errorf("not implemented")
}
func (stubStmt) ColumnConverter(int) driver.ValueConverter { return driver.DefaultParameterConverter }

type stubRows struct{}

func (stubRows) Columns() []string         { return nil }
func (stubRows) Close() error              { return nil }
func (stubRows) Next([]driver.Value) error { return io.EOF }

func TestUpdateBatchRollsBackOnNotFound(t *testing.T) {
	state := &stubState{
		execResults: []stubExecResult{
			{rowsAffected: 1},
			{rowsAffected: 0},
		},
	}
	repo := NewRepository(openStubDB(t, state))

	err := repo.UpdateBatch(context.Background(), []userdomain.User{
		{UserID: "u-1"},
		{UserID: "u-2"},
	})
	if err != userdomain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if state.beginCalls != 1 {
		t.Fatalf("expected BeginTx to be called once, got %d", state.beginCalls)
	}
	if state.commitCalls != 0 {
		t.Fatalf("expected Commit not to be called, got %d", state.commitCalls)
	}
	if state.rollbackCalls != 1 {
		t.Fatalf("expected Rollback to be called once, got %d", state.rollbackCalls)
	}
}

func TestDeleteBatchAllowsDuplicateIDs(t *testing.T) {
	state := &stubState{
		execResults: []stubExecResult{
			{rowsAffected: 1},
		},
	}
	repo := NewRepository(openStubDB(t, state))

	err := repo.DeleteBatch(context.Background(), []string{"u-1", "u-1"})
	if err != nil {
		t.Fatalf("delete batch: %v", err)
	}
	if len(state.execArgCounts) != 1 || state.execArgCounts[0] != 1 {
		t.Fatalf("expected delete to execute with 1 unique arg, got %v", state.execArgCounts)
	}
}
