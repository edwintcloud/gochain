package blockchain

// TxInput represents an input transaction.
type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

// TxOutput represents an output transaction.
type TxOutput struct {
	Value  int
	PubKey string
}

// CanUnlock verifies that data can be unlocked by the output
// that is referenced inside the TxInput.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlocked verifies that account (data) owns the
// information inside the TxOutput.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
