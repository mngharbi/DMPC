package config

import (
	"github.com/mngharbi/DMPC/core"
	"github.com/mngharbi/DMPC/users"
	"log"
	"time"
)

/*
	Helpers
*/
func buildDaemonConfig() *Config {
	confCopy := defaultDaemonConfig
	return &confCopy
}

func ensureCleanState() {
	DeleteDmpcDir()
	MakeDmpcDir()
}

/*
	Install interaction
*/

func getCliWantToReinstall() bool {
	return cliConfirm("Would you like to re-install DMPC (you will lose any current configuration)?")
}

func getCliImportingRootUser() bool {
	return !cliConfirm("Would you like to use the default properties for the root user?")
}

func getCliRootUserPath() string {
	return cliGetFilePath("Path to file containing user object you would to use:")
}

func getCliImportingKeys() bool {
	return !cliConfirm("Would you like to generate new keys?")
}

func getCliRootUser() *users.UserObject {
	userId := cliGetString("Enter the root user id you would like to use:")
	userObject := defaultUserObject
	userObject.CreatedAt = time.Now()
	userObject.UpdatedAt = time.Now()
	userObject.Id = userId
	return &userObject
}

func saveRootUserObject(userObject *users.UserObject) string {
	encoded, err := userObject.Encode()
	if err != nil {
		MakeBadStateFile()
		log.Fatalf("Unabled to encode root user object.")
		return ""
	}

	if err := WriteFile(encoded, RootUserFilename); err != nil {
		MakeBadStateFile()
		log.Fatalf("Failed to save root user file. err=%v", err)
		return ""
	}

	return GetInstallPath(RootUserFilename)
}

func getCliKeysPath(keyType string) (string, string) {
	public := cliGetFilePath("Enter path to public " + keyType + " key:")
	private := cliGetFilePath("Enter path to private " + keyType + " key:")
	return public, private
}

func getCliEncryptionKeysPath() (string, string) {
	return getCliKeysPath("encryption")
}

func getCliSigningKeysPath() (string, string) {
	return getCliKeysPath("signing")
}

func generateAndSaveKeys(isEncryption bool) (string, string) {
	var baseFilename string = SigningKeyFilename
	if isEncryption {
		baseFilename = EncryptionKeyFilename
	}

	// Make directory containing keys
	MkdirAll(KeysDir)

	// Save private key to file
	priv := core.GeneratePrivateKey()
	privString := core.PrivateAsymKeyToString(priv)
	if err := WriteFile([]byte(privString), KeysDir, baseFilename); err != nil {
		MakeBadStateFile()
		log.Fatalf("Failed to save private key file. err=%v", err)
	}

	// Save public key to file
	public := &priv.PublicKey
	publicString := core.PublicAsymKeyToString(public)
	publicFilename := baseFilename + PublicKeySuffix
	if err := WriteFile([]byte(publicString), KeysDir, publicFilename); err != nil {
		MakeBadStateFile()
		log.Fatalf("Failed to save public key file. err=%v", err)
	}

	// Return paths
	return GetInstallPath(KeysDir, publicFilename), GetInstallPath(KeysDir, baseFilename)
}

func generateAndSaveEncryptionKeys() (string, string) {
	return generateAndSaveKeys(true)
}

func generateAndSaveSigningKeys() (string, string) {
	return generateAndSaveKeys(false)
}

func saveConfig(conf *Config) {
	encoded, err := conf.Encode()
	if err != nil {
		MakeBadStateFile()
		log.Fatalf("Failed to encode configuration. err=%v", err)
	}
	if err := WriteFile(encoded, ConfigFilename); err != nil {
		MakeBadStateFile()
		log.Fatalf("Failed to save configuration file. err=%v", err)
	}
}

/*
	Main install function
*/
func Install() {
	// Check previous install
	if IsInstalled() {
		if !getCliWantToReinstall() {
			return
		}
	}

	// Ensure everything is in a clean state
	ensureCleanState()

	// Prompt for configuration
	conf := buildDaemonConfig()

	// Build root user (except keys)
	isImportingRootUser := getCliImportingRootUser()
	if isImportingRootUser {
		conf.Paths.RootUserFilePath = getCliRootUserPath()
	} else {
		newRootUser := getCliRootUser()
		conf.Paths.RootUserFilePath = saveRootUserObject(newRootUser)
	}

	// Build root user keys
	isImportingKeys := getCliImportingKeys()
	if isImportingKeys {
		conf.Paths.PublicEncryptionKeyPath, conf.Paths.PrivateEncryptionKeyPath = getCliEncryptionKeysPath()
		conf.Paths.PublicSigningKeyPath, conf.Paths.PrivateSigningKeyPath = getCliSigningKeysPath()
	} else {
		conf.Paths.PublicEncryptionKeyPath, conf.Paths.PrivateEncryptionKeyPath = generateAndSaveEncryptionKeys()
		conf.Paths.PublicSigningKeyPath, conf.Paths.PrivateSigningKeyPath = generateAndSaveSigningKeys()
	}

	saveConfig(conf)
}
