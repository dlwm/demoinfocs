package demoinfocs

import (
	"bytes"
	"testing"

	"github.com/markus-wa/demoinfocs-golang/v6/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v6/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v6/pkg/demoinfocs/msg"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCheckpointCommandsSeedWithoutDuplicateEvents(t *testing.T) {
	for _, mode := range []UserCmdParsingMode{UserCmdParsingButtonsOnly, UserCmdParsingFull} {
		t.Run(string(rune('0'+mode)), func(t *testing.T) {
			config := DefaultParserConfig
			config.UserCmdParsing = mode
			p := NewParserWithConfig(bytes.NewReader(make([]byte, 256*1024)), config).(*parser)
			defer p.Close()
			player := &common.Player{}
			p.gameState.playersByEntityID[1] = player
			commands, buttons := 0, 0
			p.RegisterEventHandler(func(events.UserCmd) { commands++ })
			p.RegisterEventHandler(func(events.PlayerButtonsStateUpdate) { buttons++ })
			full := &msg.CSVCMsg_UserCommands{Commands: []*msg.CMsgServerUserCmd{{PlayerSlot: proto.Int32(0), CmdNumber: proto.Int32(10), Data: bytesField(1, bytesField(3, varintField(1, 512)))}}}
			p.handleCheckpointUserCommands(&checkpointUserCommands{full})
			require.EqualValues(t, 512, player.ButtonsPressedState)
			require.True(t, player.ButtonsStateAvailable)
			require.Zero(t, commands)
			require.Zero(t, buttons)
			delta := &msg.CSVCMsg_UserCommands{Commands: []*msg.CMsgServerUserCmd{{PlayerSlot: proto.Int32(0), CmdNumber: proto.Int32(11), DeltaData: bytesField(1, bytesField(3, varintField(1, 1)))}}}
			p.handleUserCommands(delta)
			require.EqualValues(t, 1, player.ButtonsPressedState)
			require.Equal(t, 1, buttons)
			if mode == UserCmdParsingFull {
				require.Equal(t, 1, commands)
			} else {
				require.Zero(t, commands)
			}
			// A late or stale checkpoint cannot roll back the current command stream.
			p.handleCheckpointUserCommands(&checkpointUserCommands{full})
			require.EqualValues(t, 1, player.ButtonsPressedState)
			require.Equal(t, 1, buttons)
		})
	}
}
