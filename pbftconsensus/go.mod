module pbftconsensus

go 1.16

require (
	pbft v0.0.0-00010101000000-000000000000
	qblock v0.0.0-00010101000000-000000000000
	qbtx v0.0.0-00010101000000-000000000000
	qkdserv v0.0.0-00010101000000-000000000000
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
