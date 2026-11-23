package subvol

type SubvolSend struct {
	Sequence SendSequence
	Subvol   Summary
}

type SubvolSendList []SubvolSend

type SendItem struct {
	ParentSnapshot *Summary
	Snapshot       Summary
}

type SendSequence []SendItem
