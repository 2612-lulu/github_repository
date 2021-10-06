module qblock

go 1.16

replace (
	merkletree => ../merkletree
	qbtx => ../qbtx
	qkdserv => ../qkdserv
	uss => ../uss
	utils => ../utils
)

require (
	merkletree v0.0.0-00010101000000-000000000000
	qbtx v0.0.0-00010101000000-000000000000
	utils v0.0.0-00010101000000-000000000000
)
