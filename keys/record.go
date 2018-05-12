package keys

/*
	Record of a key
*/
type keyRecord struct {
	Id          string
	Key []byte
}

func (rec *keyRecord) Less(index string, than interface{}) bool {
	switch index {
	case "id":
		return rec.Id < than.(*keyRecord).Id
	}
	return false
}

/*
	Indexing
*/
const (
	recordIdIndex string = "id"
)
var indexesMap map[string]bool = map[string]bool{
	recordIdIndex: true,
}
func getIndexes() (res []string) {
	for k := range indexesMap {
		res = append(res, k)
	}
	return res
}
