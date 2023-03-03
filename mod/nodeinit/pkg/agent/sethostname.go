// go:build !darwin
package agent

func Sethostname(hostname []byte) error {
	//return syscall.Sethostname(hostname)
	return nil
}
