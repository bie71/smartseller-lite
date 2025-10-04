package db

import "strings"

// UnsupportedAuthPluginError indicates the MySQL server requested an authentication plugin
// that the Go MySQL driver cannot satisfy (e.g. auth_gssapi_client).
type UnsupportedAuthPluginError struct {
	Plugin string
	Err    error
}

func (e *UnsupportedAuthPluginError) Error() string {
	if e.Plugin == "" {
		return "mysql server requested an unsupported authentication plugin"
	}
	return "mysql server requested unsupported authentication plugin: " + e.Plugin
}

func (e *UnsupportedAuthPluginError) Unwrap() error {
	return e.Err
}

func classifyAuthError(err error) error {
	if err == nil {
		return nil
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "unknown auth plugin") || strings.Contains(lower, "authentication plugin is not supported") {
		plugin := extractPluginName(err.Error())
		return &UnsupportedAuthPluginError{Plugin: plugin, Err: err}
	}

	return err
}

func extractPluginName(msg string) string {
	msg = strings.TrimSpace(msg)
	// common formats:
	//   "unknown auth plugin: auth_gssapi_client"
	//   "unknown auth plugin:auth_gssapi_client"
	//   "this authentication plugin is not supported"
	if idx := strings.LastIndex(msg, ":"); idx >= 0 && idx < len(msg)-1 {
		plugin := strings.TrimSpace(msg[idx+1:])
		if plugin != "" && !strings.Contains(plugin, " ") {
			return plugin
		}
	}
	return ""
}
