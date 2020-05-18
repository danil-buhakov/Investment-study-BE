package main

func main() {
	acc, err := InitAccount()
	if err != nil {
		panic(err)
	}

	c, err := InitContact(acc)
	if err != nil {
		panic(err)
	}

	api := &API{}
	api.Init(c, acc)
}
