// go:build darwin
package agent

func Sethostname(_ []byte) error { return nil }
