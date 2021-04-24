package gormjqdt

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var (
	getRelations *RelationMappings
	queryBetween = ""
)

// New creates a new gormjqdt (it's like class constructor)
//    @params config:
//    Required params are `Model` and `Engine`
//    Model is represent the table model
//    Engine is represent the gorm DB context
func New(config ...Config) (Config, error) {
	// Set default config
	cfg := configDefault(config...)

	// Return error when model is null
	if cfg.Model == nil {
		return cfg, errors.New("[201 - GORMJQDT ERROR] - Model cannot be null")
	}
	if cfg.Engine == nil {
		return cfg, errors.New("[202 - GORMJQDT ERROR] - DB engine cannot be null")
	}

	return cfg, nil
}

// Simple to proccess server side pagination with simple approach
func (cfg Config) Simple(request RequestString, dest interface{}) (resp Response, err error) {
	var total int64
	var totalFiltered int64
	req := ParsingRequest(request)
	columns := GetDbColumns(cfg.Model)
	columnTypes := GetDbColumnTypes(cfg.Model)

	// Handle error
	err = _errorHandling(req, columns, columnTypes)
	if err != nil {
		return
	}

	// Increment draw request for response draw
	draw := req.Draw
	draw++
	resp.Draw = draw

	// Query building here
	cfg.Engine.Scopes(
		cfg.limit(*req),
		cfg.globalFilter(*req, columns),
		cfg.individualFilter(*req, columns),
		cfg.spesificFilter(*req, columns, columnTypes),
		cfg.order(*req, columns),
	).Find(dest)
	resp.Data = dest

	// Count filtered record
	cfg.Engine.Model(cfg.Model).Scopes(
		cfg.globalFilter(*req, columns),
		cfg.individualFilter(*req, columns),
		cfg.spesificFilter(*req, columns, columnTypes),
	).Count(&totalFiltered)
	resp.RecordsFiltered = totalFiltered

	// Count total record
	err = cfg.Engine.Model(cfg.Model).Count(&total).Error
	if err != nil {
		return
	}
	resp.RecordsTotal = total

	return
}

// Simple to proccess server side pagination with simple approach
func (cfg Config) Complex(
	request RequestString,
	dest interface{},
	relations map[int]interface{},
) (resp Response, err error) {
	var total int64
	var totalFiltered int64
	req := ParsingRequest(request)
	columns := GetDbColumns(cfg.Model)
	columnTypes := GetDbColumnTypes(cfg.Model)

	// Realtion setter
	getRelations = RelationSetters(relations)

	// Handle error
	err = _errorHandling(req, columns, columnTypes)
	if err != nil {
		return
	}

	// Increment draw request for response draw
	draw := req.Draw
	draw++
	resp.Draw = draw

	// Query building here
	cfg.Engine.Scopes(
		cfg.joins(*req, *getRelations),
		cfg.limit(*req),
		cfg.globalFilter(*req, columns),
		cfg.individualFilter(*req, columns),
		cfg.spesificFilter(*req, columns, columnTypes),
		cfg.order(*req, columns),
	).Find(dest)
	resp.Data = dest

	// Count filtered record
	cfg.Engine.Model(cfg.Model).Scopes(
		cfg.joins(*req, *getRelations),
		cfg.individualFilter(*req, columns),
		cfg.spesificFilter(*req, columns, columnTypes),
	).Count(&totalFiltered)
	resp.RecordsFiltered = totalFiltered

	// Count total record
	err = cfg.Engine.Model(cfg.Model).Scopes(cfg.joins(*req, *getRelations)).Count(&total).Error
	if err != nil {
		return
	}
	resp.RecordsTotal = total

	return
}

// Scopes query to filtering the data by request search (global search) jQuery DataTable
func (cfg Config) globalFilter(req ParsedRequest, columns map[int]string) func(db *gorm.DB) *gorm.DB {
	globalSearchQuery := ""

	return func(db *gorm.DB) *gorm.DB {
		if len(req.GlobalSearch) > 0 {
			for i := 0; ; i++ {
				clientColumnDataKey := fmt.Sprintf("columns[%d][data]", i)
				if req.Columns[clientColumnDataKey] == nil {
					if i == 0 {
						continue
					} else {
						break
					}
				}

				clientColumnSearchableKey := fmt.Sprintf("columns[%d][searchable]", i)
				clientColumnSearchableValue := req.Columns[clientColumnSearchableKey]

				// Determine is using regex
				clientColumnSearchRegexKey := fmt.Sprintf("columns[%d][search][regex]", i)
				clientColumnSearchRegexValue, err := strconv.ParseBool(req.Columns[clientColumnSearchRegexKey].(string))
				if err != nil {
					clientColumnSearchRegexValue = false
				}

				ok, _ := StringInArraySimple(req.Columns[clientColumnDataKey].(string), columns)
				if clientColumnSearchableValue == "true" && ok {
					// === Comon global search
					column := req.Columns[clientColumnDataKey].(string)
					column = fmt.Sprintf(`"%v"."%v"`, cfg._getTablename(db), column)
					column = cfg._castColumn(column)

					// Regex and unRegex query binding
					query := ""
					if clientColumnSearchRegexValue {
						query += cfg._bindQueryRegex(column, req.GlobalSearch)
					} else {
						query += cfg._bindQuery(column, req.GlobalSearch)
					}

					if globalSearchQuery != "" && query != "" {
						globalSearchQuery += " OR "
					}

					globalSearchQuery += query
				} else if !ok && clientColumnSearchableValue == "true" && getRelations != nil {
					// === Relations gloabl search
					column := req.Columns[clientColumnDataKey].(string)
					// Pluck the last string from clientColumnDataKey
					if strings.Contains(column, "__") {
						column = strings.Split(column, "__")[1]
					}

					// Loop through the given realtions
					for k, v := range getRelations.ModelSchemaNames {
						ok, _ := StringInArraySimple(column, getRelations.DbColumns[k])
						// Skip if key param not match with the relation table columns
						if !ok {
							continue
						}

						// Prepare column
						column = fmt.Sprintf(`"%v"."%v"`, v, column)
						column = cfg._castColumn(column)

						// Regex and unRegex query binding
						query := ""
						if clientColumnSearchRegexValue {
							query += cfg._bindQueryRegex(column, req.GlobalSearch)
						} else {
							query += cfg._bindQuery(column, req.GlobalSearch)
						}

						if globalSearchQuery != "" && query != "" {
							globalSearchQuery += " OR "
						}

						globalSearchQuery += query
					}
				} else {
					if !ok && clientColumnSearchableValue == "true" {
						log.Printf("[101 - GORMJQDT WARNING]: Something weird here. InfoTrace: %v", clientColumnDataKey)
					}
				}
			}
		}

		return db.Where(globalSearchQuery)
	}
}

// Scopes query to filtering the data by request column search (individual search) jQuery DataTable
func (cfg Config) individualFilter(req ParsedRequest, columns map[int]string) func(db *gorm.DB) *gorm.DB {
	individualSearchQuery := ""

	return func(db *gorm.DB) *gorm.DB {
		for i := 0; ; i++ {
			clientColumnDataKey := fmt.Sprintf("columns[%d][data]", i)
			if req.Columns[clientColumnDataKey] == nil {
				if i == 0 {
					continue
				} else {
					break
				}
			}

			clientColumnSearchableKey := fmt.Sprintf("columns[%d][searchable]", i)
			clientColumnSearchableValue := req.Columns[clientColumnSearchableKey]
			clientColumnSearchValKey := fmt.Sprintf("columns[%d][search][value]", i)
			clientColumnSearchValValue := req.Columns[clientColumnSearchValKey]

			// Determine is using regex
			clientColumnSearchRegexKey := fmt.Sprintf("columns[%d][search][regex]", i)
			clientColumnSearchRegexValue, err := strconv.ParseBool(req.Columns[clientColumnSearchRegexKey].(string))
			if err != nil {
				clientColumnSearchRegexValue = false
			}

			ok, _ := StringInArraySimple(req.Columns[clientColumnDataKey].(string), columns)
			if (clientColumnSearchValValue != nil && clientColumnSearchValValue != "") && clientColumnSearchableValue == "true" && ok {
				// === Comon individual search
				column := req.Columns[clientColumnDataKey].(string)
				column = fmt.Sprintf(`"%v"."%v"`, db.Statement.Table, column)
				column = cfg._castColumn(column)

				// Regex and unRegex query binding
				query := ""
				if clientColumnSearchRegexValue {
					query += cfg._bindQueryRegex(column, clientColumnSearchValValue.(string))
				} else {
					query += cfg._bindQuery(column, clientColumnSearchValValue.(string))
				}

				if individualSearchQuery != "" && query != "" {
					individualSearchQuery += " AND "
				}

				individualSearchQuery += query
			} else if (clientColumnSearchValValue != nil && clientColumnSearchValValue != "") && clientColumnSearchableValue == "true" && ok && getRelations != nil {
				// === Relation individual search
				column := req.Columns[clientColumnDataKey].(string)
				// Pluck the last string from clientColumnDataKey
				if strings.Contains(column, "__") {
					column = strings.Split(column, "__")[1]
				}

				// Loop through the given realtions
				for k, v := range getRelations.ModelSchemaNames {
					ok, _ := StringInArraySimple(column, getRelations.DbColumns[k])
					// Skip if key param not match with the relation table columns
					if !ok {
						continue
					}

					// Prepare column
					column = fmt.Sprintf(`"%v"."%v"`, v, column)
					column = cfg._castColumn(column)

					// Regex and unRegex query binding
					query := ""
					if clientColumnSearchRegexValue {
						query += cfg._bindQueryRegex(column, clientColumnSearchValValue.(string))
					} else {
						query += cfg._bindQuery(column, clientColumnSearchValValue.(string))
					}

					if individualSearchQuery != "" && query != "" {
						individualSearchQuery += " AND "
					}

					individualSearchQuery += query
				}
			} else {
				if !ok && clientColumnSearchableValue == "true" && getRelations == nil {
					log.Printf("[102 - GORMJQDT WARNING]: Something weird here. InfoTrace: %v", clientColumnDataKey)
				}
			}
		}

		return db.Where(individualSearchQuery)
	}
}

// Scopes query to filtering the data by request column search (individual search) jQuery DataTable
func (cfg Config) spesificFilter(req ParsedRequest, columns map[int]string, columnTypes map[string]reflect.Kind) func(db *gorm.DB) *gorm.DB {
	var existInDBColumn bool
	var requestedColumnKey string
	spesificSearchQuery := ""

	return func(db *gorm.DB) *gorm.DB {
		for _, item := range req.SpesificParams {
			query := ""

			switch item["key"].(type) {
			case string:
				var parsingSpesificKey string
				requestedColumnKey = item["key"].(string)
				if strings.Contains(requestedColumnKey, "~~") {
					parsingSpesificKey = strings.Split(requestedColumnKey, "~~")[0]
				} else {
					parsingSpesificKey = requestedColumnKey
				}

				existInDBColumn, _ = StringInArraySimple(parsingSpesificKey, columns)
			default:
				existInDBColumn = false
			}

			// If request spesific params not in DB column, skip that
			if !existInDBColumn && getRelations == nil {
				log.Printf("[103 - GORMJQDT WARNING]: Something weird here. InfoTrace: %v not available in DB column", requestedColumnKey)
				continue
			} else if !existInDBColumn && getRelations != nil {
				for k := range getRelations.ModelSchemaNames {
					query = cfg._bindQuerySpesific(
						fmt.Sprintf(`"%v"."%v"`, getRelations.ModelSchemaNames[k], requestedColumnKey),
						item["value"],
						getRelations.DbColumnTypes[k],
					)
				}
			} else {
				query = cfg._bindQuerySpesific(requestedColumnKey, item["value"], columnTypes)
			}

			if spesificSearchQuery != "" && query != "" {
				spesificSearchQuery += " AND "
			}

			spesificSearchQuery += query
		}

		return db.Where(spesificSearchQuery)
	}
}

// Scopes query to order the data by request order column jQuery DataTable
func (cfg Config) order(req ParsedRequest, columns map[int]string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Orders["order[0][column]"] != nil {
			for i := 0; ; i++ {
				clientColumnOrderKey := fmt.Sprintf("order[%d][column]", i)
				if req.Orders[clientColumnOrderKey] == nil {
					if i == 0 {
						continue
					} else {
						break
					}
				}

				orderedColumnIndex, err := strconv.Atoi(req.Orders["order[0][column]"].(string))
				if err != nil {
					break
				}

				clientColumnOrderableKey := fmt.Sprintf("columns[%d][orderable]", orderedColumnIndex)
				clientColumnOrderableValue := req.Columns[clientColumnOrderableKey]
				clientColumnDataKey := fmt.Sprintf("columns[%d][data]", orderedColumnIndex)
				if req.Columns[clientColumnDataKey] == nil {
					break
				}

				ok, _ := StringInArraySimple(req.Columns[clientColumnDataKey].(string), columns)
				if clientColumnOrderableValue == "true" && ok {
					clientColumnOrderDirKey := fmt.Sprintf("order[%d][dir]", i)
					clientColumnOrderDirValue := req.Orders[clientColumnOrderDirKey]

					queryOrder := fmt.Sprintf("%s %s", req.Columns[clientColumnDataKey].(string), clientColumnOrderDirValue)
					db = db.Order(queryOrder)
				} else {
					if !ok && clientColumnOrderableValue == "true" {
						log.Printf("[104 - GORMJQDT WARNING]: Something weird here. InfoTrace: columnData = %v, orderable = %v:%v",
							clientColumnDataKey,
							clientColumnOrderableKey,
							clientColumnOrderableValue,
						)
					}
				}
			}
		}

		return db
	}
}

// Scopes query to limit the data by request start and length jQuery DataTable
func (cfg Config) limit(req ParsedRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(req.Start).Limit(req.Length)
	}
}

// Scopes query to join the table inn given model to table in teh given relations parameter
func (cfg Config) joins(req ParsedRequest, relations RelationMappings) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		for _, v := range relations.ModelSchemaNames {
			db.Joins(v)
		}

		return db
	}
}

// Error handler
func _errorHandling(req *ParsedRequest, columns map[int]string, columnTypes map[string]reflect.Kind) error {
	var errTrace string

	switch true {
	case req == nil:
		errTrace = fmt.Sprintf("[GORMDTT - Error] Something wrong. ErrTrace: Request: %v", req)

	case columns == nil:
		errTrace = fmt.Sprintf("[GORMDTT - Error] Something wrong. ErrTrace: Columns: %v", columns)

	case columnTypes == nil:
		errTrace = fmt.Sprintf("[GORMDTT - Error] Something wrong. ErrTrace: ColumnTypes: %v", columnTypes)
	}

	if errTrace != "" {
		return errors.New(errTrace)
	}

	return nil
}

// Bind query but using regex
func (cfg Config) _bindQueryRegex(column string, value string) string {
	if !cfg.CaseSensitiveFilter {
		value = strings.ToLower(value)
	}

	switch cfg.Dialect {
	// pgsql
	case "postgres":
		if !cfg.CaseSensitiveFilter {
			column = fmt.Sprintf("%s ~ '%s'", column, value)
		} else {
			column = fmt.Sprintf("%s ~* '%s'", column, value)
		}

	// oracle
	case "oracle":
		column = fmt.Sprintf("REGEXP_LIKE(%s, '%s')", column, value)

	// common sql query
	default:
		column = fmt.Sprintf("%s REGEXP '%s'", column, value)
	}

	return column
}

// Bind query
func (cfg Config) _bindQuery(column string, value string) string {
	if !cfg.CaseSensitiveFilter {
		value = strings.ToLower(value)
	}

	return fmt.Sprintf("%s LIKE %s", column, "'%"+value+"%'")
}

// Bind query with spesific column
func (cfg Config) _bindQuerySpesific(column string, value interface{}, columnTypes map[string]reflect.Kind) string {
	columnKey := column
	queryColumn := column

	// Column checker
	if strings.Contains(column, ".") {
		columnKey = strings.Split(column, ".")[1]
		columnKey = strings.Replace(columnKey, `"`, "", -1)
	}

	// Safe interface conversion
	unboxedValue, isArray := ParamsValuesProcessing(value)

	// Begin more specify the meaning of params
	skip, parsed := cfg._parseMoreSpesificParamMeans(column, unboxedValue)
	if skip {
		return parsed
	}
	// End more specify the meaning of params

	switch columnTypes[columnKey] {
	// If type is integer family
	case reflect.Int,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if !isArray {
			realVal, err := strconv.Atoi(unboxedValue)
			if err != nil {
				return ""
			}

			return fmt.Sprintf("%s = %d", queryColumn, realVal)
		}

		return fmt.Sprintf("%s IN %s", queryColumn, strings.ReplaceAll(unboxedValue, "'", ""))

	// If type is decimal family
	case reflect.Float32, reflect.Float64:
		if !isArray {
			realVal, err := strconv.ParseFloat(unboxedValue, 64)
			if err != nil {
				return ""
			}

			return fmt.Sprintf("%s = %f", queryColumn, realVal)
		}

		return fmt.Sprintf("%s IN %s", queryColumn, strings.ReplaceAll(unboxedValue, "'", ""))

	// If type is boolean
	case reflect.Bool:
		realVal, err := strconv.ParseBool(unboxedValue)
		value = "NOT"
		if err == nil || realVal {
			value = ""
		}

		return fmt.Sprintf("%s IS %s TRUE", queryColumn, unboxedValue)

	// If type is string
	case reflect.String:
		if !isArray {
			if !cfg.CaseSensitiveFilter {
				queryColumn = fmt.Sprintf("LOWER(%s)", queryColumn)
				unboxedValue = strings.ToLower(unboxedValue)
			}

			return fmt.Sprintf("%s = '%s'", queryColumn, unboxedValue)
		}

		return fmt.Sprintf("%s IN %s", queryColumn, unboxedValue)

	// Maybe UUID, Blob or anything else
	default:
		if !isArray {
			return fmt.Sprintf("%s = '%s'", queryColumn, unboxedValue)
		}

		return fmt.Sprintf("%s IN %s", queryColumn, unboxedValue)
	}
}

func (cfg Config) _parseMoreSpesificParamMeans(column string, value string) (skip bool, parsed string) {
	reflectValue := reflect.ValueOf(cfg.Model).MethodByName("TableName")
	var parentTableName string

	if reflectValue.String() == "<invalid Value>" {
		parentTableName = reflect.ValueOf(cfg.Model).String()
	} else {
		parentTableName = reflectValue.Call([]reflect.Value{})[0].String()
	}

	switch true {
	case strings.Contains(column, "~~"):
		skip = true
		parsedSplit := strings.Split(column, "~~")
		column = fmt.Sprintf(`"%v"."%v"`, parentTableName, parsedSplit[0])
		operator := ""
		if parsedSplit[1] == "start" {
			operator += ">="
		} else {
			operator += "<="
		}
		if queryBetween == "" {
			queryBetween += fmt.Sprintf("(%v %v '%v'", column, operator, value)
		} else {
			queryBetween += fmt.Sprintf(" AND %v %v '%v')", column, operator, value)
			parsed = queryBetween
			queryBetween = ""
		}
	}

	return
}

// Casting DB column type based on connected DB dialect
func (cfg Config) _castColumn(column string) string {
	switch cfg.Dialect {
	case "postgres":
		column = fmt.Sprintf(`CAST(%s as TEXT)`, column)
	case "firebird":
		column = fmt.Sprintf(`CAST(%s as VARCHAR(255))`, column)
	}

	// If case sensitive is false, lower all value
	if !cfg.CaseSensitiveFilter {
		column = fmt.Sprintf("LOWER(%s)", column)
	}

	return column
}

func (cfg Config) _getTablename(db *gorm.DB) string {
	if db.Statement.Table != "" || len(db.Statement.Table) > 0 {
		return db.Statement.Table
	}

	stmt := &gorm.Statement{DB: db}
	stmt.Parse(cfg.Model)

	return stmt.Schema.Table
}
