package utils

import(
    "strconv"
    "strings"
)

func InferValue(s string) any {
    if i, err := strconv.Atoi(s); err == nil {
        return i
    }
    if f, err := strconv.ParseFloat(s, 64); err == nil {
        return f
    }
    if s == "true" {
        return true
    }
    if s == "false" {
        return false
    }
    if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
        inner := strings.Trim(s, "[]")
        if inner == "" {
            return []any{}
        }
        parts := strings.Split(inner, ",")
        arr := make([]any, 0, len(parts))
        for _, p := range parts {
            arr = append(arr, InferValue(strings.TrimSpace(p)))
        }
        return arr
    }
    return s
}
