//go:generate mockgen -source=$GOFILE -package=mock_$GOPACKAGE -destination=../mock/$GOPACKAGE/$GOFILE
package pkg

import (
	"compress/gzip"
	"io"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

var (
	devMacPattern = regexp.MustCompile(`(?P<device>[a-z0-9.]+_hwaddr)=(?P<addr>(?:[0-9A-F]{2}:){5}[0-9A-F]{2})`)
)

type IParser interface {
	BackupFile(file string) error
	ReadFile(file string) ([]byte, error)
	WriteFile(file, data string) error
	FindDeviceAddresses(data string) map[string]string
	GetMappings(source, dest map[string]string) map[string]string
	RemapDevices(source string, deviceMap map[string]string) string
}

type CfgParser struct {
	Fs afero.Fs
}

func (p CfgParser) BackupFile(file string) error {
	source, err := p.Fs.Open(file)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := p.Fs.Create(file + ".bak")
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

func (p CfgParser) ReadFile(file string) ([]byte, error) {
	f, err := p.Fs.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p CfgParser) WriteFile(file, data string) error {
	f, err := p.Fs.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := gzip.NewWriter(f)
	defer writer.Close()

	_, err = writer.Write([]byte(data))
	return err
}

func (p CfgParser) FindDeviceAddresses(data string) map[string]string {
	addresses := map[string]string{}
	matches := devMacPattern.FindAllStringSubmatch(data, -1)
	for _, match := range matches {
		addresses[match[1]] = match[2]
	}
	return addresses
}

func (c CfgParser) GetMappings(source, dest map[string]string) map[string]string {
	res := map[string]string{}
	for device, addr := range source {
		newAddr := dest[device]
		if newAddr == "" {
			continue
		}
		if res[addr] == "" {
			res[addr] = newAddr
		}
	}
	return res
}

func (c CfgParser) RemapDevices(source string, addrMap map[string]string) string {
	for orig, new := range addrMap {
		source = strings.ReplaceAll(source, orig, new)
	}
	return source
}
