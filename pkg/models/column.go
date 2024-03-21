package models

import (
	"log/slog"
	"math/big"
	"strconv"
	"time"

	"github.com/pingcap/tidb/ast"
	"github.com/pingcap/tidb/mysql"
	"github.com/pingcap/tidb/types"
)

// FunctionCall represents a MySQL function call
type FunctionCall string

// Column represents a MySQL column definition
type Column struct {
	Name             string
	Type             string
	TypeSize         int
	TypePrecision    int
	Default          any
	HasDefault       bool
	Nullable         bool
	AutoIncrementing bool
	ZeroFilled       bool
	Binary           bool
	Unsigned         bool
	InPrimaryKey     bool
	InUniqueKey      bool
	Comment          string
	Elements         []string
}

func (c *Column) setTypeInfo(tp *types.FieldType) {
	c.Type = types.TypeToStr(tp.Tp, tp.Charset)
	c.TypeSize = tp.Flen
	c.TypePrecision = tp.Decimal
	if tp.Tp == mysql.TypeEnum {
		c.Elements = tp.Elems
	}
}

func (c *Column) setFlags(flags uint) {
	c.Unsigned = mysql.HasUnsignedFlag(flags)
	c.ZeroFilled = mysql.HasZerofillFlag(flags)
	c.Binary = mysql.HasBinaryFlag(flags)
	c.AutoIncrementing = mysql.HasAutoIncrementFlag(flags)
	c.Nullable = !mysql.HasNotNullFlag(flags)
}

func (c *Column) setOptions(col *ast.ColumnDef) error {
	for _, opt := range col.Options {
		switch opt.Tp {
		case ast.ColumnOptionDefaultValue:
			c.HasDefault = true
			switch v := opt.Expr.(type) {
			case *ast.ValueExpr:
				if v != nil && !v.IsNull() {
					if err := c.setDefaultValue(col, v); err != nil {
						return err
					}
				}
			default:
				// We can't convert this type yet, so just expose the expression
				c.Default = opt.Expr
			}
		case ast.ColumnOptionNotNull:
			c.Nullable = false
		case ast.ColumnOptionAutoIncrement:
			c.AutoIncrementing = true
		case ast.ColumnOptionComment:
			c.Comment = opt.Expr.GetValue().(string)
		case ast.ColumnOptionPrimaryKey:
			c.InPrimaryKey = true
			c.InUniqueKey = true
		case ast.ColumnOptionUniqKey:
			c.InUniqueKey = true
		default:
			// Ignore other options
			slog.Warn("Unknown column option", slog.Int("type", int(opt.Tp)))
		}
	}
	return nil
}

func (c *Column) setDefaultValue(col *ast.ColumnDef, v *ast.ValueExpr) (err error) {
	if v.IsNull() {
		return nil
	}

	// We mostly ignore errors because we are only parsing - not validating. If you write
	// invalid schemas then schema2go should try to do the right thing, or at least the least-wrong thing.
	switch col.Tp.EvalType() {
	case types.ETDatetime:
		var t types.Time
		if col.Tp.Tp == mysql.TypeDate {
			t, _ = types.ParseDate(nil, v.GetString())
			c.Default, err = t.Time.GoTime(time.UTC)
			if types.ErrInvalidTimeFormat.Equal(err) {
				c.Default = time.Time{}
				return nil
			}
			c.Default = v.GetString()
			return nil
		}

		t, _ = types.ParseDatetime(nil, v.GetString())
		c.Default, err = t.Time.GoTime(time.UTC)
		if types.ErrInvalidTimeFormat.Equal(err) {
			c.Default = time.Time{}
			return nil
		}
		c.Default = v.GetString()
		return err
	case types.ETTimestamp:
		t, _ := types.ParseTimestamp(nil, v.GetString())
		c.Default, _ = t.Time.GoTime(time.UTC)
		return nil
	case types.ETDuration:
		c.Default = v.GetMysqlDuration().Duration
		if c.Default.(time.Duration) == 0 {
			d, _ := types.ParseDuration(v.GetString(), types.GetFsp(v.GetString()))
			c.Default = d.Duration
		}
	case types.ETDecimal, types.ETReal:
		switch v.Kind() {
		case types.KindFloat32:
			c.Default = v.GetFloat32()
		case types.KindFloat64:
			c.Default = v.GetFloat64()
		case types.KindMysqlDecimal:
			d := v.GetMysqlDecimal()
			prec, _ := d.PrecisionAndFrac()
			c.Default, _, err = big.ParseFloat(string(d.ToString()), 10, uint(prec), big.ToNearestEven)
		case types.KindString:
			c.Default, err = strconv.ParseFloat(v.GetString(), 64)
		}
		return err
	case types.ETInt:
		bi := big.NewInt(0)
		bi, ok := bi.SetString(v.GetString(), 10)
		if ok {
			if bi.IsInt64() {
				c.Default = bi.Int64()
			} else if bi.IsUint64() {
				c.Default = bi.Uint64()
			} else {
				c.Default = bi
			}
			return nil
		}
		c.Default = v.GetInt64()
	case types.ETJson:
		c.Default = v.GetMysqlJSON().String()
	case types.ETString:
		c.Default = v.GetString()
	default:
		c.Default = v.GetValue()
	}
	return err
}
