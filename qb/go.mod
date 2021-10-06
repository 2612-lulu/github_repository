module qb

go 1.16

require (
	github.com/boltdb/bolt v1.3.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210920023735-84f357641f63
	pbft v0.0.0-00010101000000-000000000000
	qblock v0.0.0-00010101000000-000000000000
	qbtx v0.0.0-00010101000000-000000000000
	qkdserv v0.0.0-00010101000000-000000000000
	uss v0.0.0-00010101000000-000000000000
	utils v0.0.0-00010101000000-000000000000
)

replace (
	merkletree => ../merkletree
	pbft => ../pbft
	qblock => ../qblock
	qbtx => ../qbtx
	qkdserv => ../qkdserv
	uss => ../uss
	utils => ../utils
)
