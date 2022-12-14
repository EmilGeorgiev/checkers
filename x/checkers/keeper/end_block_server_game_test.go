package keeper_test

import (
	"testing"
	"time"

	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestForfeitUnplayed(t *testing.T) {
	_, keeper, context, ctrl, _ := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	ctx := sdk.UnwrapSDKContext(context)
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "-1",
		FifoTailIndex: "-1",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeitOlderUnplayed(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Black:   carol,
		Red:     alice,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)

	nextGame, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        3,
		FifoHeadIndex: "2",
		FifoTailIndex: "2",
	}, nextGame)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeit2OldestUnplayedIn1Call(t *testing.T) {
	msgServer, keeper, context, ctrl, _ := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Black:   carol,
		Red:     alice,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: carol,
		Black:   alice,
		Red:     bob,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	game2, found := keeper.GetStoredGame(ctx, "2")
	require.True(t, found)
	game2.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game2)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)
	_, found = keeper.GetStoredGame(ctx, "2")
	require.False(t, found)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        4,
		FifoHeadIndex: "3",
		FifoTailIndex: "3",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
			{Key: "game-index", Value: "2"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeitPlayedOnce(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	pay := escrow.ExpectPay(context, bob, 45).Times(1)
	escrow.ExpectRefund(context, bob, 45).Times(1).After(pay)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "-1",
		FifoTailIndex: "-1",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeitOlderPlayedOnce(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Red:     carol,
		Black:   alice,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        3,
		FifoHeadIndex: "2",
		FifoTailIndex: "2",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "*"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeit2OldestPlayedOnceIn1Call(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Black:   carol,
		Red:     alice,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "2",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: carol,
		Black:   alice,
		Red:     bob,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	game1.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game1)
	game2, found := keeper.GetStoredGame(ctx, "2")
	require.True(t, found)
	game2.Deadline = types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	keeper.SetStoredGame(ctx, game2)
	keeper.ForfeitExpiredGames(context)

	_, found = keeper.GetStoredGame(ctx, "1")
	require.False(t, found)
	_, found = keeper.GetStoredGame(ctx, "2")
	require.False(t, found)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        4,
		FifoHeadIndex: "3",
		FifoTailIndex: "3",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t,
		sdk.StringEvent{
			Type: "game-forfeited",
			Attributes: []sdk.Attribute{
				{Key: "game-index", Value: "1"},
				{Key: "winner", Value: "*"},
				{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
				{Key: "game-index", Value: "2"},
				{Key: "winner", Value: "*"},
				{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*"},
			},
		}, event)
}

func TestForfeitPlayedTwice(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	oldDeadline := types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	game1.Deadline = oldDeadline
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	game1, found = keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "",
		Turn:        "b",
		Black:       bob,
		Red:         carol,
		MoveCount:   uint64(2),
		BeforeIndex: "-1",
		AfterIndex:  "-1",
		Deadline:    oldDeadline,
		Winner:      "r",
		Wager:       45,
	}, game1)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "-1",
		FifoTailIndex: "-1",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "r"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeitOlderPlayedTwice(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Black:   carol,
		Red:     alice,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	oldDeadline := types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	game1.Deadline = oldDeadline
	keeper.SetStoredGame(ctx, game1)
	keeper.ForfeitExpiredGames(context)

	game1, found = keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "",
		Turn:        "b",
		Black:       bob,
		Red:         carol,
		MoveCount:   uint64(2),
		BeforeIndex: "-1",
		AfterIndex:  "-1",
		Deadline:    oldDeadline,
		Winner:      "r",
		Wager:       45,
	}, game1)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        3,
		FifoHeadIndex: "2",
		FifoTailIndex: "2",
	}, systemInfo)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "r"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}

func TestForfeit2OldestPlayedTwiceIn1Call(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	ctx := sdk.UnwrapSDKContext(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: bob,
		Black:   carol,
		Red:     alice,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "2",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   alice,
		GameIndex: "2",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: carol,
		Black:   alice,
		Red:     bob,
	})
	game1, found := keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	oldDeadline := types.FormatDeadline(ctx.BlockTime().Add(time.Duration(-1)))
	game1.Deadline = oldDeadline
	keeper.SetStoredGame(ctx, game1)
	game2, found := keeper.GetStoredGame(ctx, "2")
	require.True(t, found)
	game2.Deadline = oldDeadline
	keeper.SetStoredGame(ctx, game2)
	keeper.ForfeitExpiredGames(context)

	game1, found = keeper.GetStoredGame(ctx, "1")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "1",
		Board:       "",
		Turn:        "b",
		Black:       bob,
		Red:         carol,
		MoveCount:   uint64(2),
		BeforeIndex: "-1",
		AfterIndex:  "-1",
		Deadline:    oldDeadline,
		Winner:      "r",
		Wager:       45,
	}, game1)

	game2, found = keeper.GetStoredGame(ctx, "2")
	require.True(t, found)
	require.EqualValues(t, types.StoredGame{
		Index:       "2",
		Board:       "",
		Turn:        "b",
		Black:       carol,
		Red:         alice,
		MoveCount:   uint64(2),
		BeforeIndex: "-1",
		AfterIndex:  "-1",
		Deadline:    oldDeadline,
		Winner:      "r",
	}, game2)

	systemInfo, found := keeper.GetSystemInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        4,
		FifoHeadIndex: "3",
		FifoTailIndex: "3",
	}, systemInfo)

	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 3)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-forfeited",
		Attributes: []sdk.Attribute{
			{Key: "game-index", Value: "1"},
			{Key: "winner", Value: "r"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*"},
			{Key: "game-index", Value: "2"},
			{Key: "winner", Value: "r"},
			{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*"},
		},
	}, event)
}