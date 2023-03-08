package cmd

import (
	"testing"

	"github.com/cybozu-go/log"
	"github.com/golang/mock/gomock"
	"github.com/hsn723/tomato-migrate/mock/pkg"
	"github.com/hsn723/tomato-migrate/pkg"
	"github.com/stretchr/testify/assert"
)

func setupMockParser(t *testing.T, f func(m *mock_pkg.MockIParser)) *mock_pkg.MockIParser {
	t.Helper()

	ctrl := gomock.NewController(t)
	parser := mock_pkg.NewMockIParser(ctrl)
	f(parser)
	return parser
}

func TestSetup(t *testing.T) {
	cases := []struct {
		title   string
		verbose bool
	}{
		{
			title: "Default",
		},
		{
			title:   "Verbose",
			verbose: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			verbose = tc.verbose
			err := setup(nil, nil)
			assert.NoError(t, err)
			_, ok := parser.(pkg.CfgParser)
			assert.True(t, ok)
			actual := log.DefaultLogger()
			if tc.verbose {
				assert.Equal(t, log.LvInfo, actual.Threshold())
			} else {
				assert.Equal(t, log.LvError, actual.Threshold())
			}
		})
	}
}

func TestRunRoot(t *testing.T) {
	inFile = "dummy_in.cfg"
	outFile = "dummy_out.cfg"
	cases := []struct {
		title string
		setup func(m *mock_pkg.MockIParser)
		isErr bool
	}{
		{
			title: "Success",
			setup: func(m *mock_pkg.MockIParser) {
				gomock.InOrder(
					m.EXPECT().BackupFile(gomock.Eq(outFile)),
					m.EXPECT().ReadFile(gomock.Eq(outFile)).Return([]byte("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n"), nil),
					m.EXPECT().FindDeviceAddresses(gomock.Eq("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n")).Return(map[string]string{
						"wan_hwaddr": "8E:E5:6F:48:48:1D",
					}),
					m.EXPECT().ReadFile(gomock.Eq(inFile)).Return([]byte("TCF1\x14\x00\x00\x00wan_hwaddr=57:98:6D:BD:55:94\x00\x00\n"), nil),
					m.EXPECT().FindDeviceAddresses(gomock.Eq("TCF1\x14\x00\x00\x00wan_hwaddr=57:98:6D:BD:55:94\x00\x00\n")).Return(map[string]string{
						"wan_hwaddr": "57:98:6D:BD:55:94",
					}),
					m.EXPECT().GetMappings(gomock.Eq(map[string]string{
						"wan_hwaddr": "57:98:6D:BD:55:94",
					}), gomock.Eq(map[string]string{
						"wan_hwaddr": "8E:E5:6F:48:48:1D",
					})).Return(map[string]string{
						"57:98:6D:BD:55:94": "8E:E5:6F:48:48:1D",
					}),
					m.EXPECT().RemapDevices(gomock.Eq("TCF1\x14\x00\x00\x00wan_hwaddr=57:98:6D:BD:55:94\x00\x00\n"), gomock.Eq(map[string]string{
						"57:98:6D:BD:55:94": "8E:E5:6F:48:48:1D",
					})).Return("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n"),
					m.EXPECT().WriteFile(gomock.Eq(outFile), gomock.Eq("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n")),
				)
			},
		},
		{
			title: "BackupFailure",
			setup: func(m *mock_pkg.MockIParser) {
				m.EXPECT().BackupFile(gomock.Eq(outFile)).Return(assert.AnError)
			},
			isErr: true,
		},
		{
			title: "ReadDestFailure",
			setup: func(m *mock_pkg.MockIParser) {
				m.EXPECT().BackupFile(gomock.Eq(outFile))
				m.EXPECT().ReadFile(gomock.Eq(outFile)).Return(nil, assert.AnError)
			},
			isErr: true,
		},
		{
			title: "ReadSourceFailure",
			setup: func(m *mock_pkg.MockIParser) {
				m.EXPECT().BackupFile(gomock.Eq(outFile))
				m.EXPECT().ReadFile(gomock.Eq(outFile)).Return([]byte("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n"), nil)
				m.EXPECT().FindDeviceAddresses(gomock.Eq("TCF1\x14\x00\x00\x00wan_hwaddr=8E:E5:6F:48:48:1D\x00\x00\n")).Return(map[string]string{
					"wan_hwaddr": "8E:E5:6F:48:48:1D",
				})
				m.EXPECT().ReadFile(gomock.Eq(inFile)).Return(nil, assert.AnError)
			},
			isErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			t.Helper()
			parser = setupMockParser(t, tc.setup)
			err := runRoot(rootCmd, nil)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
