package uploaders

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-deploy-to-bitrise-io/bundletool"
)

// DeployAAB ...
func DeployAAB(pth, buildURL, token, notifyUserGroups, notifyEmails, isEnablePublicPage string) (string, error) {
	log.Printf("- analyzing aab")

	tmpPth, err := pathutil.NormalizedOSTempDirPath("aab-bundle")
	if err != nil {
		return "", err
	}

	r, err := bundletool.NewRunner()
	if err != nil {
		return "", err
	}

	// create debug keystore for signing
	log.Printf("- generating debug keystore")
	keystorePath := filepath.Join(tmpPth, "debug.keystore")
	if out, err := command.New("keytool", "-genkey", "-v", "-keystore", keystorePath, "-storepass", "android", "-alias", "androiddebugkey", "-keypass", "android", "-keyalg", "RSA", "-keysize", "2048", "-validity", "10000", "-dname", "C=US, O=Android, CN=Android Debug").RunAndReturnTrimmedCombinedOutput(); err != nil {
		return "", errors.Wrap(err, out)
	}

	// generate `tmpDir/universal.apks` from aab file
	log.Printf("- generating universal apk")
	apksPth := filepath.Join(tmpPth, "universal.apks")
	if out, err := r.Execute("build-apks", "--mode=universal", "--bundle", pth, "--output", apksPth, "--ks", keystorePath, "--ks-pass", "pass:android", "--ks-key-alias", "androiddebugkey", "--key-pass", "pass:android"); err != nil {
		return "", errors.Wrap(err, out)
	}

	// unzip `tmpDir/universal.apks` to tmpPth to have `tmpDir/universal.apk`
	if out, err := command.New("unzip", "-v", apksPth, "-d", tmpPth).RunAndReturnTrimmedCombinedOutput(); err != nil {
		return "", errors.Wrap(err, out)
	}

	// rename `tmpDir/universal.apk` to `tmpDir/aab-name.aab.universal.apk`
	universalAPKPath := filepath.Join(tmpPth, "universal.apk")
	renamedUniversalAPKPath := filepath.Join(tmpPth, filepath.Base(pth)+".universal.apk")
	if err := os.Rename(universalAPKPath, renamedUniversalAPKPath); err != nil {
		return "", err
	}

	// get aab manifest dump
	out, err := r.Execute("dump", "manifest", "--bundle", pth)
	if err != nil {
		return "", errors.Wrap(err, out)
	}

	packageName, versionCode, versionName := filterPackageInfos(out)

	appInfo := map[string]interface{}{
		"package_name": packageName,
		"version_code": versionCode,
		"version_name": versionName,
	}

	log.Printf("  aab infos: %v", appInfo)

	if packageName == "" {
		log.Warnf("Package name is undefined, AndroidManifest.xml package content:\n%s", out)
	}

	if versionCode == "" {
		log.Warnf("Version code is undefined, AndroidManifest.xml package content:\n%s", out)
	}

	if versionName == "" {
		log.Warnf("Version name is undefined, AndroidManifest.xml package content:\n%s", out)
	}

	// ---

	fileSize, err := fileSizeInBytes(pth)
	if err != nil {
		return "", fmt.Errorf("failed to get apk size, error: %s", err)
	}

	aabInfoMap := map[string]interface{}{
		"file_size_bytes": fmt.Sprintf("%f", fileSize),
		"app_info":        appInfo,
	}

	artifactInfoBytes, err := json.Marshal(aabInfoMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal apk infos, error: %s", err)
	}

	// ---

	uploadURL, artifactID, err := createArtifact(buildURL, token, pth, "android-apk")
	if err != nil {
		return "", fmt.Errorf("failed to create apk artifact, error: %s", err)
	}

	if err := uploadArtifact(uploadURL, pth, ""); err != nil {
		return "", fmt.Errorf("failed to upload apk artifact, error: %s", err)
	}

	if _, err = finishArtifact(buildURL, token, artifactID, string(artifactInfoBytes), "", "", "false"); err != nil {
		return "", fmt.Errorf("failed to finish apk artifact, error: %s", err)
	}

	fmt.Println()

	return DeployAPK(renamedUniversalAPKPath, buildURL, token, notifyUserGroups, notifyEmails, isEnablePublicPage)
}
