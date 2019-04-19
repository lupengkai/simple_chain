package main

type TXOutput struct {
	Value int
	ScriptPubKey string
}

func (output TXOutput) CanBeUnlockedWith(s string) bool {

}
