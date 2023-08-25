package validation

import (
	"fmt"
	"reflect"
)

func ValidateStructFields(structPtr any) error {
	const (
		tagKey   = "required"
		tagValue = "true"
	)
	if err := validateStructPtr(structPtr); err != nil {
		return fmt.Errorf("validate struct pointer error: %v", err)
	}
	refVal, refType := structPtrReflection(structPtr)

	for fieldIdx := 0; fieldIdx < refVal.NumField(); fieldIdx++ {
		var (
			tagVal string
		)
		if tagVal = refType.Field(fieldIdx).Tag.Get(tagKey); tagVal == tagValue {
			var (
				fieldVal reflect.Value
			)
			if fieldVal = refVal.Field(fieldIdx); !fieldVal.IsValid() || fieldVal.IsZero() {
				return fmt.Errorf("not specified value for struct field \"%s\" with tag `%s:\"%s\"`",
					refType.Field(fieldIdx).Name, tagKey, tagValue)
			}
			if fieldIface := fieldVal.Interface(); validateStructPtr(fieldIface) == nil {
				return ValidateStructFields(fieldIface)
			}
		}
	}
	return nil
}

func structPtrReflection(structPtr any) (reflect.Value, reflect.Type) {
	refVal := reflect.ValueOf(structPtr).Elem()
	refType := reflect.TypeOf(structPtr).Elem()
	return refVal, refType
}

func validateStructPtr(structPtr any) error {
	if refStructPtr := reflect.ValueOf(structPtr); refStructPtr.Kind() == reflect.Ptr {
		if refStruct := refStructPtr.Elem(); refStruct.Kind() == reflect.Struct {
			return nil
		}
		return fmt.Errorf("function argument not a struct pointer")
	}
	return fmt.Errorf("function argument not a pointer")
}
