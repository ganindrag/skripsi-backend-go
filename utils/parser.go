package utils

import (
	"fmt"
	"reflect"
	"strings"
	"github.com/go-playground/validator/v10"
)

func ParseValidator(errors validator.ValidationErrors) (errmap map[string]string) {
	errmap = map[string]string{}
    for _, err := range errors {
		switch err.Tag() {
			case "required":
				errmap[err.Field()] = fmt.Sprintf("Field '%s' cannot be blank!", err.Field())
			case "email":
				errmap[err.Field()] = fmt.Sprintf("Field '%s' must be email!", err.Field())
			default:
				errmap[err.Field()] = fmt.Sprintf("Something wrong with field '%s'", err.Field())
		}
	}
	return errmap
}

func ParseStructToUpdateSql(typeof reflect.Type, valueof reflect.Value) (sqlUpdate string, paramsUpdate []interface{}) {
	var stringBuilder strings.Builder
    valueLen := valueof.NumField()
    paramAt := 1
    for i := 0; i < valueLen; i++ {
    	_, ignored := typeof.Field(i).Tag.Lookup("pgxignored")
    	fmt.Println(typeof.Field(i).Name, reflect.Zero(typeof))
    	fmt.Println(valueof.Field(i))
    	if !valueof.Field(i).IsZero() && !ignored {
	    	if stringBuilder.Len() != 0 {
	    		stringBuilder.WriteString(", ")
	    	}
	    	fieldName := typeof.Field(i).Name
	    	pgxcolumn, ok := typeof.Field(i).Tag.Lookup("pgxcolumn")
	    	if ok {
	    		fieldName = pgxcolumn
	    	}
	    	stringBuilder.WriteString(fmt.Sprintf("%s=$%d", strings.ToLower(fieldName), paramAt))
	    	paramsUpdate = append(paramsUpdate, valueof.Field(i).Interface())
	    	paramAt++
    	}
    }
    sqlUpdate = stringBuilder.String()
    return
}

func ParseStructToInsertSql(typeof reflect.Type, valueof reflect.Value) (columnsInsert string, paramsInsert string, valuesInsert []interface{}) {
	var sbColumns, sbParams strings.Builder
    valueLen := valueof.NumField()
    paramAt := 1
    for i := 0; i < valueLen; i++ {
    	_, ignored := typeof.Field(i).Tag.Lookup("pgxignored")
    	if !valueof.Field(i).IsZero() && !ignored {
	    	if sbColumns.Len() != 0 {
	    		sbColumns.WriteString(", ")
	    		sbParams.WriteString(", ")
	    	}
	    	fieldName := typeof.Field(i).Name
	    	pgxcolumn, ok := typeof.Field(i).Tag.Lookup("pgxcolumn")
	    	if ok {
	    		fieldName = pgxcolumn
	    	}
	    	sbColumns.WriteString(strings.ToLower(fieldName))
	    	sbParams.WriteString(fmt.Sprintf("$%d", paramAt))
	    	valuesInsert = append(valuesInsert, valueof.Field(i).Interface())
	    	paramAt++
    	}
    }
    columnsInsert = sbColumns.String()
    paramsInsert = sbParams.String()
    return
}