package main

func main() {
	mainCmd := CreateEchoCmd()

	e := mainCmd.Execute()
	if e != nil {
		panic(e)
	}
}
