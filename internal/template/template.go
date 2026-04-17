package template

import "strings"

// Render 渲染模板
func Render(data interface{}, params map[string]string) interface{} {
	switch v := data.(type) {
	case string:
		result := v
		for k, val := range params {
			result = strings.ReplaceAll(result, "{{.PathParams."+k+"}}", val)
		}
		return result

	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = Render(val, params)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = Render(val, params)
		}
		return result

	default:
		return data
	}
}
