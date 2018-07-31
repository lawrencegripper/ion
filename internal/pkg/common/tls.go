package common

// TLSCerts for mutual auth
type TLSCerts struct {
	CertFile   string
	KeyFile    string
	CACertFile string
}

// Available returns whether all required certs have been provided
func (t *TLSCerts) Available() bool {
	return (t.CACertFile != "" && t.CertFile != "" && t.KeyFile != "")
}
