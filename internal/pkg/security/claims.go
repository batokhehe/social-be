package security

import "fmt"

func GetIntClaim(claims map[string]interface{}, key string) (int, error) {
	raw, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("missing %s claim", key)
	}

	value, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid %s claim type", key)
	}

	return int(value), nil
}

func GetStringClaim(claims map[string]interface{}, key string) (string, error) {
	raw, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing %s claim", key)
	}

	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("invalid %s claim type", key)
	}

	return value, nil
}
