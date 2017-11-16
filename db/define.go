// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance wdbh the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/abcum/surreal/sql"
	"github.com/abcum/surreal/util/keys"
	"github.com/abcum/surreal/util/rand"
)

func (e *executor) executeDefineNamespace(ctx context.Context, ast *sql.DefineNamespaceStatement) (out []interface{}, err error) {

	nkey := &keys.NS{KV: ast.KV, NS: ast.Name.ID}
	_, err = e.dbo.Put(0, nkey.Encode(), ast.Encode())

	return

}

func (e *executor) executeDefineDatabase(ctx context.Context, ast *sql.DefineDatabaseStatement) (out []interface{}, err error) {

	e.dbo.AddNS(ast.NS)
	dkey := &keys.DB{KV: ast.KV, NS: ast.NS, DB: ast.Name.ID}
	_, err = e.dbo.Put(0, dkey.Encode(), ast.Encode())

	return

}

func (e *executor) executeDefineLogin(ctx context.Context, ast *sql.DefineLoginStatement) (out []interface{}, err error) {

	ast.Code = rand.New(128)

	ast.Pass, _ = bcrypt.GenerateFromPassword(ast.Pass, bcrypt.DefaultCost)

	switch ast.Kind {
	case sql.NAMESPACE:
		e.dbo.AddNS(ast.NS)
		ukey := &keys.NU{KV: ast.KV, NS: ast.NS, US: ast.User.ID}
		_, err = e.dbo.Put(0, ukey.Encode(), ast.Encode())
	case sql.DATABASE:
		e.dbo.AddDB(ast.NS, ast.DB)
		ukey := &keys.DU{KV: ast.KV, NS: ast.NS, DB: ast.DB, US: ast.User.ID}
		_, err = e.dbo.Put(0, ukey.Encode(), ast.Encode())
	}

	return

}

func (e *executor) executeDefineToken(ctx context.Context, ast *sql.DefineTokenStatement) (out []interface{}, err error) {

	switch ast.Kind {
	case sql.NAMESPACE:
		e.dbo.AddNS(ast.NS)
		tkey := &keys.NT{KV: ast.KV, NS: ast.NS, TK: ast.Name.ID}
		_, err = e.dbo.Put(0, tkey.Encode(), ast.Encode())
	case sql.DATABASE:
		e.dbo.AddDB(ast.NS, ast.DB)
		tkey := &keys.DT{KV: ast.KV, NS: ast.NS, DB: ast.DB, TK: ast.Name.ID}
		_, err = e.dbo.Put(0, tkey.Encode(), ast.Encode())
	}

	return

}

func (e *executor) executeDefineScope(ctx context.Context, ast *sql.DefineScopeStatement) (out []interface{}, err error) {

	ast.Code = rand.New(128)

	e.dbo.AddDB(ast.NS, ast.DB)

	skey := &keys.SC{KV: ast.KV, NS: ast.NS, DB: ast.DB, SC: ast.Name.ID}
	_, err = e.dbo.Put(0, skey.Encode(), ast.Encode())

	return

}

func (e *executor) executeDefineTable(ctx context.Context, ast *sql.DefineTableStatement) (out []interface{}, err error) {

	e.dbo.AddDB(ast.NS, ast.DB)

	for _, TB := range ast.What {

		ast.Name = sql.NewIdent(TB.TB)

		tkey := &keys.TB{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: TB.TB}
		if _, err = e.dbo.Put(0, tkey.Encode(), ast.Encode()); err != nil {
			return nil, err
		}

		if ast.Lock {

			for _, FT := range ast.From {
				tkey := &keys.FT{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: FT.TB, FT: TB.TB}
				if _, err = e.dbo.Put(0, tkey.Encode(), ast.Encode()); err != nil {
					return nil, err
				}
			}

		}

	}

	return

}

func (e *executor) executeDefineEvent(ctx context.Context, ast *sql.DefineEventStatement) (out []interface{}, err error) {

	for _, TB := range ast.What {

		e.dbo.AddTB(ast.NS, ast.DB, TB.TB)

		ekey := &keys.EV{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: TB.TB, EV: ast.Name.ID}
		if _, err = e.dbo.Put(0, ekey.Encode(), ast.Encode()); err != nil {
			return nil, err
		}

	}

	return

}

func (e *executor) executeDefineField(ctx context.Context, ast *sql.DefineFieldStatement) (out []interface{}, err error) {

	for _, TB := range ast.What {

		e.dbo.AddTB(ast.NS, ast.DB, TB.TB)

		fkey := &keys.FD{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: TB.TB, FD: ast.Name.ID}
		if _, err = e.dbo.Put(0, fkey.Encode(), ast.Encode()); err != nil {
			return nil, err
		}

	}

	return

}

func (e *executor) executeDefineIndex(ctx context.Context, ast *sql.DefineIndexStatement) (out []interface{}, err error) {

	for _, TB := range ast.What {

		e.dbo.AddTB(ast.NS, ast.DB, TB.TB)

		ikey := &keys.IX{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: TB.TB, IX: ast.Name.ID}
		if _, err = e.dbo.Put(0, ikey.Encode(), ast.Encode()); err != nil {
			return nil, err
		}

		dkey := &keys.Index{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: TB.TB, IX: ast.Name.ID, FD: keys.Ignore}
		if _, err = e.dbo.DelP(0, dkey.Encode(), 0); err != nil {
			return nil, err
		}

		if _, err = e.executeUpdate(ctx, &sql.UpdateStatement{What: []sql.Expr{TB}}); err != nil {
			return nil, err
		}

	}

	return

}
