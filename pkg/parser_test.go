package pkg

import (
	_ "embed"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var (
	//go:embed t/tomato_v128_input
	sampleInputFile []byte
)

func TestBackupFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title      string
		fsContents []string
		file       string
		isErr      bool
	}{
		{
			title:      "Success",
			fsContents: []string{"sample.cfg"},
			file:       "sample.cfg",
		},
		{
			title: "NoSource",
			file:  "sample.cfg",
			isErr: true,
		},
		{
			title: "Overwrite",
			fsContents: []string{
				"sample.cfg",
				"sample.cfg.bak",
			},
			file: "sample.cfg",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			fs := afero.NewMemMapFs()
			for _, fsContent := range tc.fsContents {
				_, err := fs.Create(fsContent)
				assert.NoError(t, err)
			}
			parser := CfgParser{Fs: fs}
			err := parser.BackupFile(tc.file)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				actual, err := afero.Exists(fs, tc.file+".bak")
				assert.NoError(t, err)
				assert.True(t, actual)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		file   string
		expect []byte
		isErr  bool
	}{
		{
			title:  "Success",
			file:   "t/sample.cfg",
			expect: []byte("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n"),
		},
		{
			title: "NoFile",
			file:  "dummy",
			isErr: true,
		},
		{
			title: "NotGz",
			file:  "t/tomato_v128_input",
			isErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			fs := afero.NewOsFs()
			parser := CfgParser{Fs: fs}
			actual, err := parser.ReadFile(tc.file)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expect, actual)
		})
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title string
		file  string
		data  string
		isErr bool
	}{
		{
			title: "Success",
			file:  "output.cfg",
			data:  "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			fs := afero.NewMemMapFs()
			parser := CfgParser{Fs: fs}
			err := parser.WriteFile(tc.file, tc.data)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindDeviceAddresses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		data   string
		expect map[string]string
	}{
		{
			title:  "BlankFile",
			expect: map[string]string{},
		},
		{
			title: "Success",
			data:  string(sampleInputFile),
			expect: map[string]string{
				"wl1_hwaddr":    "64:15:FE:FD:BE:E5",
				"wl0.14_hwaddr": "98:F3:9B:12:E0:2F",
				"wl1.12_hwaddr": "A5:BE:9E:44:00:21",
				"wl0.10_hwaddr": "57:98:6D:BD:55:94",
				"lan_hwaddr":    "E9:AF:D3:95:F0:01",
				"wl1.14_hwaddr": "29:74:3A:BE:8F:D2",
				"wan_hwaddr":    "8E:E5:6F:48:48:1D",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			parser := CfgParser{}
			actual := parser.FindDeviceAddresses(tc.data)
			assert.Equal(t, tc.expect, actual)
		})
	}
}

func TestGetMappings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		source map[string]string
		dest   map[string]string
		expect map[string]string
	}{
		{
			title: "Success",
			source: map[string]string{
				"wl1_hwaddr":    "64:15:FE:FD:BE:E5",
				"wl0.14_hwaddr": "98:F3:9B:12:E0:2F",
				"wl1.12_hwaddr": "A5:BE:9E:44:00:21",
				"lan_hwaddr":    "E9:AF:D3:95:F0:01",
				"wan_hwaddr":    "8E:E5:6F:48:48:1D",
			},
			dest: map[string]string{
				"wl1_hwaddr":    "25:51:AB:F6:81:50",
				"wl0.14_hwaddr": "0B:3A:5F:2D:A0:5C",
				"wl1.12_hwaddr": "5B:9F:6D:2F:0F:DA",
				"lan_hwaddr":    "A8:A5:54:9D:92:75",
				"wan_hwaddr":    "2B:3D:FF:B9:09:53",
			},
			expect: map[string]string{
				"64:15:FE:FD:BE:E5": "25:51:AB:F6:81:50",
				"98:F3:9B:12:E0:2F": "0B:3A:5F:2D:A0:5C",
				"A5:BE:9E:44:00:21": "5B:9F:6D:2F:0F:DA",
				"E9:AF:D3:95:F0:01": "A8:A5:54:9D:92:75",
				"8E:E5:6F:48:48:1D": "2B:3D:FF:B9:09:53",
			},
		},
		{
			title: "MissingDest",
			source: map[string]string{
				"wl1_hwaddr":    "64:15:FE:FD:BE:E5",
				"wl0.14_hwaddr": "98:F3:9B:12:E0:2F",
				"wl1.12_hwaddr": "A5:BE:9E:44:00:21",
				"lan_hwaddr":    "E9:AF:D3:95:F0:01",
				"wan_hwaddr":    "8E:E5:6F:48:48:1D",
			},
			dest: map[string]string{
				"wl1_hwaddr":    "25:51:AB:F6:81:50",
				"wl0.14_hwaddr": "0B:3A:5F:2D:A0:5C",
				"wl1.12_hwaddr": "5B:9F:6D:2F:0F:DA",
				"wan_hwaddr":    "2B:3D:FF:B9:09:53",
			},
			expect: map[string]string{
				"64:15:FE:FD:BE:E5": "25:51:AB:F6:81:50",
				"98:F3:9B:12:E0:2F": "0B:3A:5F:2D:A0:5C",
				"A5:BE:9E:44:00:21": "5B:9F:6D:2F:0F:DA",
				"8E:E5:6F:48:48:1D": "2B:3D:FF:B9:09:53",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			parser := CfgParser{}
			actual := parser.GetMappings(tc.source, tc.dest)
			assert.Equal(t, tc.expect, actual)
		})
	}
}

func TestRemapDevices(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title   string
		source  string
		addrMap map[string]string
		expect  string
	}{
		{
			title:  "Success",
			source: "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00wl0.14_hwaddr=64:15:FE:FD:BE:E5\x00hoge=hige\x00et0macaddr=E9:AF:D3:95:F0:01\x00other=82:9A:32:DE:C3:6C\x00lan_hwaddr=E9:AF:D3:95:F0:01\x00\x00\n",
			addrMap: map[string]string{
				"64:15:FE:FD:BE:E5": "25:51:AB:F6:81:50",
				"98:F3:9B:12:E0:2F": "0B:3A:5F:2D:A0:5C",
				"A5:BE:9E:44:00:21": "5B:9F:6D:2F:0F:DA",
				"E9:AF:D3:95:F0:01": "A8:A5:54:9D:92:75",
				"8E:E5:6F:48:48:1D": "2B:3D:FF:B9:09:53",
			},
			expect: "TCF1\x14\x00\x00\x00wan_hwaddr=2B:3D:FF:B9:09:53\x00wl0.14_hwaddr=25:51:AB:F6:81:50\x00hoge=hige\x00et0macaddr=A8:A5:54:9D:92:75\x00other=82:9A:32:DE:C3:6C\x00lan_hwaddr=A8:A5:54:9D:92:75\x00\x00\n",
		},
		{
			title:  "NoReplacements",
			source: "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00wl0.14_hwaddr=64:15:FE:FD:BE:E5\x00hoge=hige\x00et0macaddr=E9:AF:D3:95:F0:01\x00other=82:9A:32:DE:C3:6C\x00\x00\n",
			expect: "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00wl0.14_hwaddr=64:15:FE:FD:BE:E5\x00hoge=hige\x00et0macaddr=E9:AF:D3:95:F0:01\x00other=82:9A:32:DE:C3:6C\x00\x00\n",
		},
		{
			title:  "NoMatchingDevices",
			source: "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00wl0.14_hwaddr=64:15:FE:FD:BE:E6\x00hoge=hige\x00et0macaddr=E9:AF:D3:95:F0:02\x00other=82:9A:32:DE:C3:6C\x00\x00\n",
			addrMap: map[string]string{
				"64:15:FE:FD:BE:E5": "25:51:AB:F6:81:50",
				"A5:BE:9E:44:00:21": "5B:9F:6D:2F:0F:DA",
				"E9:AF:D3:95:F0:01": "A8:A5:54:9D:92:75",
			},
			expect: "TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00wl0.14_hwaddr=64:15:FE:FD:BE:E6\x00hoge=hige\x00et0macaddr=E9:AF:D3:95:F0:02\x00other=82:9A:32:DE:C3:6C\x00\x00\n",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			parser := CfgParser{}
			actual := parser.RemapDevices(tc.source, tc.addrMap)
			assert.Equal(t, tc.expect, actual)
		})
	}
}
