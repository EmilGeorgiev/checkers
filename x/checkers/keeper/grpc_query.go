package keeper

import (
	"github.com/EmilGeorgiev/checkers/x/checkers/types"
)

var _ types.QueryServer = Keeper{}
