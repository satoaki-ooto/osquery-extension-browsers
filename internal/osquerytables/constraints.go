package osquerytables

import (
	"sort"
	"strconv"

	"github.com/osquery/osquery-go/plugin/table"
)

func equalityStrings(queryContext table.QueryContext, column string) map[string]struct{} {
	constraintList, present := queryContext.Constraints[column]
	if !present {
		return nil
	}

	values := make(map[string]struct{})
	for _, constraint := range constraintList.Constraints {
		if constraint.Operator == table.OperatorEquals {
			values[constraint.Expression] = struct{}{}
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func equalityInt64s(queryContext table.QueryContext, column string) []int64 {
	stringValues := equalityStrings(queryContext, column)
	values := make([]int64, 0, len(stringValues))
	for value := range stringValues {
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			values = append(values, parsedValue)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values
}

func matchesEquality(value string, allowedValues map[string]struct{}) bool {
	if len(allowedValues) == 0 {
		return true
	}
	_, present := allowedValues[value]
	return present
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
