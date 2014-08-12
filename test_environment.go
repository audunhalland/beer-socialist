package tbeer

func OpenTestEnv() {
	InitDB()
	InitRestTree()
}

func CloseTestEnv() {
	GlobalDB.Close()
	GlobalDB = nil
	restTree = newSelectDP()
}
