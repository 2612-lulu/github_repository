module pbft

go 1.16

replace (
	merkletree => ../merkletree
	qblock => ../qblock
	qbtx => ../qbtx
	qkdserv => ../qkdserv
	uss => ../uss
	utils => ../utils
)

require (
	qblock v0.0.0-00010101000000-000000000000
	qbtx v0.0.0-00010101000000-000000000000
	qkdserv v0.0.0-00010101000000-000000000000
	uss v0.0.0-00010101000000-000000000000
	utils v0.0.0-00010101000000-000000000000
)
