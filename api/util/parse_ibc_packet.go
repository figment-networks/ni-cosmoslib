package util

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/figment-networks/indexing-engine/structs"
)

var bigZero = new(big.Int).SetInt64(0)

type PacketData struct {
	Amount   string `json:"amount"`
	Denom    string `json:"denom"`
	Receiver string `json:"receiver"`
	Sender   string `json:"sender"`
}

func ParsePacket(data []byte, event *structs.SubsetEvent) error {
	if len(data) == 0 {
		return nil
	}
	var packetDataInterface interface{}
	err := json.Unmarshal(data, &packetDataInterface)
	if err != nil {
		return fmt.Errorf("packet malformed: %w %s", err, string(data))
	}
	m, ok := packetDataInterface.(map[string]interface{})
	if !ok {
		return fmt.Errorf("packet malformed: %s", string(data))
	}
	if len(m) != 4 {
		return fmt.Errorf("packet malformed: %s", string(data))
	}
	_, amtSok := m["amount"]
	_, denomSok := m["denom"]
	_, receiveSok := m["receiver"]
	_, senderSok := m["sender"]
	if !(amtSok && denomSok && receiveSok && senderSok) {
		return fmt.Errorf("packet malformed: %s", string(data))
	}
	var packetData *PacketData
	err = json.Unmarshal(data, &packetData)
	if err != nil {
		return fmt.Errorf("packet malformed: %w %s", err, string(data))
	}
	amt, ok := new(big.Int).SetString(packetData.Amount, 10)
	if !ok {
		return fmt.Errorf("packet amount not a string: %v", packetData)
	}
	if amt.Cmp(bigZero) < 0 || len(packetData.Denom) == 0 || len(packetData.Sender) == 0 || len(packetData.Receiver) == 0 {
		return fmt.Errorf("packet malformed: %v", packetData)
	}
	// adding the Amount on the receiver.
	event.Sender = []structs.EventTransfer{
		{
			Account: structs.Account{ID: packetData.Sender},
		},
	}
	event.Recipient = []structs.EventTransfer{
		{
			Account: structs.Account{ID: packetData.Receiver},
			Amounts: []structs.TransactionAmount{
				{
					Text:     amt.String(),
					Numeric:  amt,
					Currency: packetData.Denom,
				},
			},
		},
	}
	return nil
}
