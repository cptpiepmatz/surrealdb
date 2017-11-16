// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
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
	"fmt"

	"context"

	"github.com/abcum/surreal/sql"
	"github.com/abcum/surreal/util/data"
	"github.com/abcum/surreal/util/keys"
)

func (e *executor) executeDelete(ctx context.Context, stm *sql.DeleteStatement) ([]interface{}, error) {

	var what sql.Exprs

	for _, val := range stm.What {
		w, err := e.fetch(ctx, val, nil)
		if err != nil {
			return nil, err
		}
		what = append(what, w)
	}

	i := newIterator(e, ctx, stm, false)

	for _, w := range what {

		switch what := w.(type) {

		default:
			return nil, fmt.Errorf("Can not execute DELETE query using value '%v'", what)

		case *sql.Table:
			key := &keys.Table{KV: stm.KV, NS: stm.NS, DB: stm.DB, TB: what.TB}
			i.processTable(ctx, key)

		case *sql.Ident:
			key := &keys.Table{KV: stm.KV, NS: stm.NS, DB: stm.DB, TB: what.ID}
			i.processTable(ctx, key)

		case *sql.Thing:
			key := &keys.Thing{KV: stm.KV, NS: stm.NS, DB: stm.DB, TB: what.TB, ID: what.ID}
			i.processThing(ctx, key)

		case *sql.Model:
			key := &keys.Thing{KV: stm.KV, NS: stm.NS, DB: stm.DB, TB: what.TB, ID: nil}
			i.processModel(ctx, key, what)

		case *sql.Batch:
			key := &keys.Thing{KV: stm.KV, NS: stm.NS, DB: stm.DB, TB: what.TB, ID: nil}
			i.processBatch(ctx, key, what)

		}

	}

	return i.Yield(ctx)

}

func (e *executor) fetchDelete(ctx context.Context, stm *sql.DeleteStatement, doc *data.Doc) (interface{}, error) {

	stm.Echo = sql.BEFORE

	if doc != nil {
		vars := data.New()
		vars.Set(doc, varKeyParent)
		ctx = context.WithValue(ctx, ctxKeySubs, vars)
	}

	out, err := e.executeDelete(ctx, stm)
	if err != nil {
		return nil, err
	}

	switch len(out) {
	case 1:
		return data.Consume(out).Get(docKeyOne, docKeyId).Data(), nil
	default:
		return data.Consume(out).Get(docKeyAll, docKeyId).Data(), nil
	}

}

func (d *document) runDelete(ctx context.Context, stm *sql.DeleteStatement) (interface{}, error) {

	var ok bool
	var err error
	var met = _DELETE

	defer d.close()

	if err = d.setup(); err != nil {
		return nil, err
	}

	if d.val.Exi() == false {
		return nil, nil
	}

	if ok, err = d.allow(ctx, met); err != nil {
		return nil, err
	} else if ok == false {
		return nil, nil
	}

	if ok, err = d.check(ctx, stm.Cond); err != nil {
		return nil, err
	} else if ok == false {
		return nil, nil
	}

	if err = d.erase(); err != nil {
		return nil, err
	}

	if err = d.purgeIndex(); err != nil {
		return nil, err
	}

	if stm.Hard {
		if err = d.eraseThing(); err != nil {
			return nil, err
		}
	} else {
		if err = d.purgeThing(); err != nil {
			return nil, err
		}
	}

	if err = d.table(ctx, met); err != nil {
		return nil, err
	}

	if err = d.event(ctx, met); err != nil {
		return nil, err
	}

	if err = d.lives(ctx, met); err != nil {
		return nil, err
	}

	return d.yield(ctx, stm, stm.Echo)

}
