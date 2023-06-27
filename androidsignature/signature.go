package androidsignature

import (
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
)

const (
	validSignatureMessage = "jar verified"
)

// Read ...
func Read(path string) (string, error) {
	cmd := signatureCheckCommand(path)
	output, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", err
	}

	var signature string

	if strings.Contains(output, validSignatureMessage) {
		// The signature details appear in the output in the following format:
		// - Signed by "C=Aa, ST=Bbbbb, L=Ccccc, O=Ddddd, OU=Eeeee, CN=Fffff"
		regex := regexp.MustCompile(`- Signed by "(.*?)"`)
		matches := regex.FindStringSubmatch(output)
		if len(matches) == 2 {
			signature = matches[1]
		}
	}

	return signature, nil
}

func signatureCheckCommand(path string) *command.Model {
	params := []string{"-verify", "-certs", "-verbose", path}
	return command.New("jarsigner", params...)
}
