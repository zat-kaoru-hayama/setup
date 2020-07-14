package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/zat-kaoru-hayama/go-msidb"
)

var kernel32 = syscall.NewLazyDLL("kernel32")
var getPrivateProfileString = kernel32.NewProc("GetPrivateProfileStringW")

func GetPrivateProfileString(path, section, keyname, defaultValue string) (string, error) {
	var buffer [2048]uint16

	sec16, secErr := syscall.UTF16PtrFromString(section)
	if secErr != nil {
		return "", secErr
	}
	key16, keyErr := syscall.UTF16PtrFromString(keyname)
	if keyErr != nil {
		return "", keyErr
	}
	default16, defaultErr := syscall.UTF16PtrFromString(defaultValue)
	if defaultErr != nil {
		return "", defaultErr
	}
	path16, pathErr := syscall.UTF16PtrFromString(path)
	if pathErr != nil {
		return "", pathErr
	}
	result, _, err := getPrivateProfileString.Call(
		uintptr(unsafe.Pointer(sec16)),
		uintptr(unsafe.Pointer(key16)),
		uintptr(unsafe.Pointer(default16)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(path16)))

	if result <= 0 {
		return "", err
	}
	return syscall.UTF16ToString(buffer[0:result]), nil
}

func callMsi(msiname string, upgrade bool) error {
	var cmd1 *exec.Cmd
	if upgrade {
		cmd1 = exec.Command(
			"msiexec", "/i", msiname, "REINSTALL=ALL", "REINSTALLMODE=vomus")
	} else {
		cmd1 = exec.Command("msiexec", "/i", msiname)
	}
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	cmd1.Stdin = os.Stdin
	return cmd1.Run()
}

func getSetupIniPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(exePath)
	setupIniPath := exePath[:len(exePath)-len(ext)] + ".ini"
	_, err = os.Stat(setupIniPath)
	if err != nil {
		return "", fmt.Errorf("%s: %w", setupIniPath, err)
	}
	return setupIniPath, nil
}

func getMsiPath() (string, error) {
	iniPath, err := getSetupIniPath()
	if err != nil {
		return "", err
	}
	value, err := GetPrivateProfileString(iniPath, "Product0", "MsiPath1041", "")
	if err == nil {
		return value, nil
	}
	return GetPrivateProfileString(iniPath, "Product0", "MsiPath", "")
}

func mains() error {
	uninit := msidb.CoInit()
	defer uninit()

	msiPath, err := getMsiPath()
	if err != nil {
		return err
	}

	db, err := msidb.Query(msiPath)
	if err != nil {
		return err
	}
	productCode, ok := db["ProductCode"]
	if !ok {
		return fmt.Errorf("%s: ProductCode not found")
	}

	isUpgrade := msidb.IsInstalled(productCode)

	return callMsi(msiPath, isUpgrade)
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
