package tinypdf

type ICacheColorText interface {
	ICacheContent
	equal(obj ICacheColorText) bool
}
