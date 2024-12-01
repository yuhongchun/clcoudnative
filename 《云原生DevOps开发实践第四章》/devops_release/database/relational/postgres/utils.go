package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	nlog "github.com/sirupsen/logrus"
)

var dsn string
var PostgresUtils = PostgresDbUtils{}

type Options struct {
	Dsn string
}
type PostgresDbUtils struct {
	Db *sqlx.DB
}

func (*PostgresDbUtils) GetDsn() string {
	return dsn
}
func (p *PostgresDbUtils) GetConnect(ctx context.Context) (*sql.Conn, error) {
	conn, err := p.Db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func (p *PostgresDbUtils) Open() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil

}
func (p *PostgresDbUtils) SetUp(o Options) {
	dsn = o.Dsn
	tempDb, err := p.Open()
	if err != nil {
		nlog.Panic("数据库连接失败！err:", err)
	}
	p.Db = tempDb
}
func (p *PostgresDbUtils) Close(db *sqlx.DB) {
	db.Close()
}
func (p *PostgresDbUtils) PrepareQuery(ctx context.Context, query string, dest interface{}, args interface{}) error {
	stmt, err := p.Db.PrepareNamedContext(ctx, query)
	defer p.CloseStmt(stmt)
	if err != nil {
		nlog.Errorf(err.Error())
		return err
	}
	// rows, err := stmt.QueryxContext(ctx, args)
	// if err != nil {
	// 	nlog.Errorf(err.Error())
	// 	return err
	// }
	err = stmt.SelectContext(ctx, dest, args)
	if err != nil {
		return err
	}
	return err
}
func (p *PostgresDbUtils) PrepareQueryRaw(ctx context.Context, query string, dest interface{}, args interface{}) error {
	stmt, err := p.Db.PrepareNamedContext(ctx, query)
	defer p.CloseStmt(stmt)
	if err != nil {
		return err
	}
	row := stmt.QueryRowxContext(ctx, args)
	err = row.StructScan(dest)
	if err != nil {
		return err
	}
	return nil
}
func (p *PostgresDbUtils) PrepareExec(ctx context.Context, exec string, args ...interface{}) (sql.Result, error) {
	stmt, err := p.Db.PrepareNamedContext(ctx, exec)
	defer p.CloseStmt(stmt)
	if err != nil {
		fmt.Println(exec)
		return nil, err
	}
	fmt.Println(args...)
	res, err := stmt.ExecContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (p *PostgresDbUtils) CloseStmt(stmt *sqlx.NamedStmt) {
	if stmt == nil {
		return
	}
	err := stmt.Close()
	if err != nil {
		nlog.Error(err)
	}
}
