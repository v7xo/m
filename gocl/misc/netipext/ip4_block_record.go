package netipext

type IP4BlockRecord struct {
	Tag   string
	Block IP4Block
}

func (v IP4BlockRecord) String() string {
	return "(" + v.Block.String() + ", tag: " + v.Tag + ")"
}

type IP4BlockRecordList []IP4BlockRecord

func (v IP4BlockRecordList) ToBlockList(sel func(IP4BlockRecord) bool) IP4BlockList {
	if v == nil {
		return nil
	}

	res := make(IP4BlockList, 0)

	for i := range v {
		if sel == nil || sel(v[i]) {
			res = append(res, v[i].Block)
		}
	}

	return res
}
