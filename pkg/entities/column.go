package entities

import (
	"log/slog"
	"math/big"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/types"
)

const (
	// TypeEnum represents a MySQL enum type. This is because ENUMS are not a real type in MySQL, they are just a list of strings.
	TypeEnum = "enum"
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
	switch {
	case tp.GetType() == mysql.TypeTiny:
		c.Type = "tinyint"
	case tp.GetType() == mysql.TypeShort:
		c.Type = "smallint"
	case tp.GetType() == mysql.TypeInt24:
		c.Type = "mediumint"
	case tp.GetType() == mysql.TypeLonglong:
		c.Type = "bigint"
	default:
		c.Type = tp.EvalType().String()
	}
	c.TypeSize = tp.GetFlen()
	c.TypePrecision = tp.GetDecimal()
	if tp.GetType() == mysql.TypeEnum {
		c.Type = TypeEnum
		c.Elements = tp.GetElems()
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
			case ast.ValueExpr:
				if v != nil && v.GetValue() != nil {
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
			c.Comment = opt.Expr.Text()
		case ast.ColumnOptionPrimaryKey:
			c.InPrimaryKey = true
			c.InUniqueKey = true
			c.Nullable = false // Primary keys are not nullable
		case ast.ColumnOptionUniqKey:
			c.InUniqueKey = true
		default:
			// Ignore other options
			slog.Warn("Unhandled column option", slog.Int("type", int(opt.Tp)))
		}
	}
	return nil
}

func (c *Column) setDefaultValue(col *ast.ColumnDef, v ast.ValueExpr) (err error) {
	if v.GetValue() == nil {
		return nil
	}

	// We mostly ignore errors because we are only parsing - not validating. If you write
	// invalid schemas then schema2go should try to do the right thing, or at least the least-wrong thing.
	switch col.Tp.EvalType() {
	case types.ETDatetime:
		if col.Tp.GetType() == mysql.TypeDate {
			c.Default, err = time.Parse(time.DateOnly, v.GetString())
			if err != nil {
				c.Default = time.Time{}
				return nil
			}

			c.Default = v.GetString()
			return nil
		}

		c.Default, err = time.Parse(time.DateTime, v.GetString())
		if err != nil {
			c.Default = time.Time{}
			return nil
		}

		c.Default = v.GetString()
		return err
	case types.ETTimestamp:
		t, err := time.Parse(time.DateTime, v.GetString())
		if err != nil {
			c.Default = time.Time{}
			return nil
		}

		c.Default = t
		return nil
	case types.ETDuration:
		d, err := time.ParseDuration(v.GetString())
		if err != nil {
			c.Default = time.Duration(0)
			return nil
		}

		c.Default = d
		return nil
	case types.ETDecimal, types.ETReal:
		switch v.GetType().GetType() {
		case mysql.TypeFloat, mysql.TypeDouble:
			c.Default = v.GetValue()
			return nil
		case mysql.TypeNewDecimal:
			d := v.GetString()
			precision := col.Tp.GetFlen()

			var prec uint
			if precision >= 0 {
				prec = uint(precision)
			}

			c.Default, _, err = big.ParseFloat(d, 10, prec, big.ToNearestEven)
		}
		return err
	case types.ETInt:
		bi := big.NewInt(0)
		bi, ok := bi.SetString(v.GetString(), 10)
		if ok {
			switch {
			case bi.IsInt64():
				c.Default = bi.Int64()
			case bi.IsUint64():
				c.Default = bi.Uint64()
			default:
				c.Default = bi
			}
			return nil
		}
		c.Default = v.GetValue()
	case types.ETJson:
		c.Default = v.GetString()
	case types.ETString:
		c.Default = v.GetString()
	default:
		c.Default = v.GetValue()
	}
	return err
}
