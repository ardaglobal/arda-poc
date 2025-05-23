package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/ardaglobal/arda-poc/x/property/types"
)

func hashProperty(p types.Property) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

type transferHashData struct {
	Msg   *types.MsgTransferShares `json:"msg"`
	Final types.Property           `json:"final"`
}

func hashTransfer(msg *types.MsgTransferShares, final types.Property) (string, error) {
	data := transferHashData{Msg: msg, Final: final}
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
